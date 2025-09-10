package common

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"unicode"

	"github.com/codeclysm/extract/v4"
	"gopkg.in/yaml.v3"
)

var g_yamlPath string = ""
var Verbose = false
var env map[string]string = map[string]string{}
var packageYML PackageYAMLSimple

type PackageYAMLSimple struct {
	Main          string              `yaml:"main,omitempty"`
	Package       string              `yaml:"package,omitempty"`
	Src           string              `yaml:"src,omitempty"`
	Version       string              `yaml:"version,omitempty"`
	Language      string              `yaml:"language,omitempty"`
	Env           string              `yaml:"env,omitempty"`
	Scripts       map[string]string   `yaml:"scripts,omitempty"`
	Dependencies  []string            `yaml:"dependencies,omitempty"`
	Repos         []map[string]string `yaml:"repos,omitempty"`
	Args          map[string]string   `yaml:"args,omitempty"`
	OtherSections map[string]any      `yaml:",inline,omitempty"`
}

// PackageYAML represents the structure of package.yml with ordered fields
type PackageYAML struct {
	Main          string         `yaml:"main,omitempty"`
	Package       string         `yaml:"package,omitempty"`
	Src           string         `yaml:"src,omitempty"`
	Version       string         `yaml:"version,omitempty"`
	Language      string         `yaml:"language,omitempty"`
	Env           string         `yaml:"env,omitempty"`
	Scripts       *OrderedMap    `yaml:"scripts,omitempty"`
	Dependencies  []string       `yaml:"dependencies"`
	Repos         []*OrderedMap  `yaml:"repos,omitempty"`
	Args          *OrderedMap    `yaml:"args,omitempty"`
	OtherSections map[string]any `yaml:",inline,omitempty"`
}

// OrderedMap represents a map that maintains insertion order
type OrderedMap struct {
	Values map[string]any
	Order  []string
}

func (om *OrderedMap) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping node, got %v", value.Kind)
	}

	om.Values = make(map[string]any)
	om.Order = make([]string, 0, len(value.Content)/2)

	for i := 0; i < len(value.Content); i += 2 {
		key := value.Content[i].Value
		var val any
		if err := value.Content[i+1].Decode(&val); err != nil {
			return err
		}
		// Only add non-empty values
		if val != nil && val != "" && val != false {
			om.Values[key] = val
			om.Order = append(om.Order, key)
		}
	}
	return nil
}

func (om *OrderedMap) MarshalYAML() (interface{}, error) {
	node := &yaml.Node{
		Kind:    yaml.MappingNode,
		Content: make([]*yaml.Node, 0, len(om.Order)*2),
	}
	for _, key := range om.Order {
		if val, exists := om.Values[key]; exists {
			// Skip empty values
			if val == nil || val == "" || val == false {
				continue
			}
			keyNode := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: key,
			}
			valNode := &yaml.Node{}
			if err := valNode.Encode(val); err != nil {
				return nil, err
			}
			node.Content = append(node.Content, keyNode, valNode)
		}
	}

	// If no non-empty values, return empty mapping node to preserve structure
	if len(node.Content) == 0 {
		return node, nil
	}
	return node, nil
}

func Init() {
	homeDir := HomeDir()
	configPath := filepath.Join(homeDir, "config.json")
	type configStruct struct {
		Verbose bool `json:"verbose"`
	}
	var cfg configStruct

	file, err := os.Open(configPath)
	if err != nil {
		// config.json does not exist, set Verbose to false
		Verbose = false
		return
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		Verbose = false
		return
	}
	Verbose = cfg.Verbose
}

func JAVAC() string {
	javac := "javac"
	if IsWindows() {
		javac = javac + ".exe"
	}
	DCEVM := filepath.Join(HomeDir(), "dcevm", "bin", javac)
	if _, err := os.Stat(DCEVM); !os.IsNotExist(err) {
		return DCEVM
	}
	return "javac"
}
func JAVA() string {
	java := "java"
	if IsWindows() {
		java = java + ".exe"
	}
	DCEVM := filepath.Join(HomeDir(), "dcevm", "bin", java)
	if _, err := os.Stat(DCEVM); !os.IsNotExist(err) {
		return DCEVM
	}
	return "java"
}

func KOTLINC() string {
	kotlinc := ""
	ktc := "kotlinc"
	if IsWindows() {
		ktc = ktc + ".bat"
	}
	kotlinc = filepath.Join(HomeDir(), "kotlinc", "bin", ktc)
	if _, err := os.Stat(kotlinc); !os.IsNotExist(err) {
		return kotlinc
	}
	return ""
}

func CheckDeps(dep string) {
	deps := GetDependencies(true)
	seen := make(map[string]bool)
	seen[NormalizeSpaces(dep)] = true
	for _, d := range deps {
		if seen[NormalizeSpaces(d)] {
			println("duplicate dependency found: " + d)
			os.Exit(1)
		}
		seen[NormalizeSpaces(d)] = true
	}
}

func VerifyPackageYML() {
	defer func() {
		if r := recover(); r != nil {
			println("package.yml has syntax error\nmake sure that lists are lists and objects are objects and values are values")
			os.Exit(1)
		}
	}()
	data, err := os.ReadFile(g_yamlPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// load in package.yml
	if err := yaml.Unmarshal(data, &packageYML); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// load in env
	loadInEnv()
	// verify duplicate deps
	CheckDeps("")
	// verify repos
	repoSection := GetSection("repos", false)
	if repoSection != nil {
		repoSectionListMap := repoSection.([]any)
		alreadyThere := []string{}
		for _, r := range repoSectionListMap {
			repoSectionMap := r.(map[string]string)
			for k, v := range repoSectionMap {
				if strings.HasPrefix(k, "-") || slices.Contains(alreadyThere, k) || k == "local" || k == "raw" || k == "jpm" || v == "" || k == "_" {
					println("syntax error with: ", k, v)
					os.Exit(1)
				}
				if k != "username" && k != "password" {
					alreadyThere = append(alreadyThere, k)
				}
			}
		}
	}

	// verify dependencies
	depSection := GetSection("dependencies", false).([]string)
	if depSection != nil {
		set := map[string]bool{}
		for _, v := range depSection {
			s := strings.Split(v, ":")
			sj := []string{}
			if len(s) > 1 {
				sj = strings.Split(s[len(s)-1], " ")
				sj = sj[1:]
				s = s[:len(s)-1]
			}
			val := strings.Join(append(s, sj...), " ")
			if _, ok := set[val]; ok {
				println("duplicate dependency:", val)
				os.Exit(1)
			}
			set[val] = true
		}
	}
}
func FindPackageYML() (string, string) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if g_yamlPath != "" {
		return g_yamlPath, filepath.Dir(g_yamlPath)
	}

	for {
		ymlPath := filepath.Join(dir, "package.yml")
		if _, err := os.Stat(ymlPath); err == nil {
			os.Chdir(dir)
			g_yamlPath = ymlPath
			VerifyPackageYML()
			return ymlPath, dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	println("package.yml not found")
	os.Exit(1)
	return "", ""
}

func ReplaceDependency(oldDepString string, newDepString string) {
	data, err := os.ReadFile(g_yamlPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var pkgYAML PackageYAML
	if err := yaml.Unmarshal(data, &pkgYAML); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for i, depStr := range pkgYAML.Dependencies {
		if NormalizeSpaces(depStr) != NormalizeSpaces(oldDepString) && NormalizeSpaces(depStr)+":" != NormalizeSpaces(oldDepString) {
			continue
		}
		pkgYAML.Dependencies[i] = NormalizeSpaces(newDepString)
		break
	}
	// Write back to file with comment preservation
	if err := WriteYAML(g_yamlPath, pkgYAML); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func AddToSection(sectionName string, sectionValue any) {
	data, err := os.ReadFile(g_yamlPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var pkgYAML PackageYAML
	if err := yaml.Unmarshal(data, &pkgYAML); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	switch sectionName {
	case "dependencies":
		if pkgYAML.Dependencies == nil {
			pkgYAML.Dependencies = make([]string, 0)
		}
		newDependency := sectionValue.(string)
		if !slices.Contains(pkgYAML.Dependencies, newDependency) {
			pkgYAML.Dependencies = append(pkgYAML.Dependencies, newDependency)
		}
	case "repos":
		if pkgYAML.Repos == nil {
			pkgYAML.Repos = []*OrderedMap{}
		}
		if repoMap, ok := sectionValue.(map[string]string); ok {
			newRepos := &OrderedMap{
				Values: map[string]any{},
				Order:  []string{},
			}

			for k, v := range repoMap {
				fmt.Println("Adding", k+":", v)
				newRepos.Values[k] = v
				newRepos.Order = append(newRepos.Order, k)
			}
			pkgYAML.Repos = append(pkgYAML.Repos, newRepos)
			//pkgYAML.Repos = append(pkgYAML.Repos, repoMap)
		}
	case "scripts":
		if pkgYAML.Scripts == nil {
			pkgYAML.Scripts = &OrderedMap{
				Values: make(map[string]any),
				Order:  make([]string, 0),
			}
		}
		if scriptMap, ok := sectionValue.(map[string]any); ok {
			for k, v := range scriptMap {
				pkgYAML.Scripts.Values[k] = v
				pkgYAML.Scripts.Order = append(pkgYAML.Scripts.Order, k)
			}
		}
	case "args":
		if pkgYAML.Args == nil {
			pkgYAML.Args = &OrderedMap{
				Values: make(map[string]any),
				Order:  make([]string, 0),
			}
		}
		if argMap, ok := sectionValue.(map[string]any); ok {
			for k, v := range argMap {
				pkgYAML.Args.Values[k] = v
				pkgYAML.Args.Order = append(pkgYAML.Args.Order, k)
			}
		}
	}

	// Write back to file with comment preservation
	if err := WriteYAML(g_yamlPath, pkgYAML); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func GetSection(section string, isFatal bool) any {
	if packageYML.Package == "" {
		println("package section of package.yml cannot be empty")
		os.Exit(1)
	}
	switch section {
	case "main":
		return ParseENV(packageYML.Main)
	case "language":
		return ParseENV(packageYML.Language)
	case "version":
		return ParseENV(packageYML.Version)
	case "env":
		return packageYML.Env
	case "package":
		return ParseENV(packageYML.Package)
	case "src":
		return ParseENV(packageYML.Src)
	case "scripts":
		scripts := map[string]string{}
		for key, v := range packageYML.Scripts {
			scripts[key] = ParseENV(v)
		}
		return scripts
	case "dependencies":
		deps := []string{}
		for _, v := range packageYML.Dependencies {
			deps = append(deps, ParseENV(v))
		}
		return deps
	case "repos":
		repos := []any{}
		for i := range packageYML.Repos {
			repoData := map[string]string{}
			for k, v := range packageYML.Repos[i] {
				repoData[k] = ParseENV(v)
			}
			repos = append(repos, repoData)
		}
		return repos
	case "args":
		args := map[string]string{}
		for key, v := range packageYML.Args {

			args[key] = ParseENV(v)
		}
		return args
	default:
		return nil
	}
}

func ParseScripts() map[string]string {
	return GetSection("scripts", true).(map[string]string)
}

func ParseArgs() map[string]string {
	val := GetSection("args", false)
	args := make(map[string]string)
	if scriptsSection, ok := val.(map[string]string); ok {
		for key, value := range scriptsSection {
			args[key] = strings.TrimSpace(value)
		}
	}
	return args
}
func loadInEnv() {
	// Attempt to read a .env file in the same directory as package.yml
	envPath := packageYML.Env
	// Handle both forward and backslashes for Windows compatibility
	envPath = filepath.Clean(envPath)
	if data, err := os.ReadFile(envPath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				env[key] = val
			}
		}
	}
	env["main"] = ParseENV(packageYML.Main)
	env["package"] = ParseENV(packageYML.Package)
	env["src"] = ParseENV(packageYML.Src)
	env["version"] = ParseENV(packageYML.Version)
	env["language"] = ParseENV(packageYML.Language)
	for k, v := range packageYML.OtherSections {
		if str, ok := v.(string); ok {
			env[k] = ParseENV(str)
		}
	}
}
func ParseENV(str string) string {
	result := str
	for {
		startIdx := strings.Index(result, "{{")
		if startIdx == -1 {
			break
		}
		endIdx := strings.Index(result[startIdx:], "}}")
		if endIdx == -1 {
			break
		}
		endIdx += startIdx
		varName := strings.TrimSpace(result[startIdx+2 : endIdx])
		val, ok := env[varName]
		if !ok {
			// If not found, remove the unmatched {{KEY}} to avoid infinite loop
			result = result[:startIdx] + result[endIdx+2:]
			continue
		}
		result = result[:startIdx] + val + result[endIdx+2:]
	}
	return result
}
func ParseEnvVars(prefix string) string {
	envString := ""
	if str := GetSection("env", false); str != "" {
		file, err := os.Open(str.(string))
		if err != nil {
			println("env file was not found :" + err.Error())
			os.Exit(1)
			return ""
		}
		defer file.Close()

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) == "" {
				continue
			}
			l := strings.SplitN(line, "=", 2)
			lines = append(lines, prefix+l[0]+"=\""+l[1]+"\"")
		}

		if err := scanner.Err(); err != nil {
			return ""
		}
		envString = strings.Join(lines, "\n") + "\n"
	}
	return envString
}

func GetDependencies(isFatal bool) []string {
	depListRawNil := GetSection("dependencies", isFatal)
	if depListRawNil == nil {
		return []string{}
	}
	depListRaw := depListRawNil.([]string)
	var deps []string = []string{}
	for _, item := range depListRaw {
		deps = append(deps, strings.TrimSpace(item))

	}
	return deps
}
func RunCMD(script string, showStdOut bool) error {
	var cmd *exec.Cmd
	if tmpFile, err := os.CreateTemp("", "jpm_script_*.cmd"); err == nil {
		tmpFile.WriteString("@echo off\n" + script)
		if Verbose {
			println("\033[33m--(verbose command)==>  " + script + "\033[0m")
			showStdOut = true
		}
		tmpFile.Close()
		cmd = exec.Command("cmd", "/C", tmpFile.Name())
		defer os.Remove(tmpFile.Name())
		if dir, err := os.Getwd(); err == nil {
			cmd.Dir = dir
		}
		if showStdOut {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
		}

		// Set up signal handling for Ctrl+C
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT)
		defer signal.Stop(sigChan)
		// Start the command
		if err := cmd.Start(); err == nil {
			done := make(chan error, 1)
			go func() {
				done <- cmd.Wait()
			}()
			select {
			case err := <-done:
				return err
			case sig := <-sigChan:
				// Kill the process immediately when Ctrl+C is pressed
				if cmd.Process != nil {
					cmd.Process.Kill()
				}
				return fmt.Errorf("process terminated by signal: %v", sig)
			}
		}
		return fmt.Errorf("failed to trap signals")
	}
	return fmt.Errorf("failed to run")
}
func RunScript(script string, showStdOut bool) error {
	if Verbose {
		println("\033[33m--(verbose command)==> " + script + "\033[0m")
		showStdOut = true
	}
	var cmd *exec.Cmd
	if IsWindows() {
		cmd = exec.Command("C:\\Program Files\\Git\\bin\\bash.exe", "-c", script)
	} else {
		cmd = exec.Command("sh", "-c", script)
	}
	if dir, err := os.Getwd(); err == nil {
		cmd.Dir = dir
	}
	if showStdOut {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
	}
	return cmd.Run()
}

func HomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory can't be determined
		return "."
	}
	appDir := ".jpm"
	// Create the directory if it doesn't exist
	dirPath := filepath.Join(homeDir, appDir)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			// If we can't create the directory, fall back to current directory
			return "."
		}
	}
	return dirPath
}

func SrcDir() string {
	valsrc := GetSection("src", false)
	if str := valsrc.(string); str != "" {
		return str
	} else {
		valpkg := GetSection("package", true)
		if str = valpkg.(string); str != "" {
			return strings.ReplaceAll(str, "-", "_")
		}
		return ""
	}
}

func NormalizeSpaces(s string) string {
	// Split by spaces and filter out empty strings
	parts := strings.Fields(s)
	// Join back with single spaces
	return strings.Join(parts, " ")
}
func Extract(dirname string, filename string) {
	file, _ := os.Open(filepath.Join(dirname, filename))
	extract.Archive(context.TODO(), file, dirname, nil)
}
func CleanupExtract(dirname string, filename string) {
	os.Remove(filepath.Join(dirname, filename))
}
func DownloadFile(url string, dirpath string, filename string, override bool, returnContent bool) (error, []byte) {
	if Verbose {
		println(url)
	}
	filePath := filepath.Join(dirpath, filename)
	// Get the data first to check response
	resp, err := http.Get(url)
	if err != nil {
		return err, nil
	}
	defer resp.Body.Close()

	if _, err := os.Stat(filePath); err == nil {
		if override {
			os.Remove(filePath)
		} else {
			return errors.New("file already exist"), nil
		}
	}

	// Check server response before creating file
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", resp.Status), nil
	}

	if returnContent {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err, nil
		}
		return nil, bodyBytes
	} else {
		// Create the file only after successful HTTP response
		out, err := os.Create(filePath)
		if err != nil {
			return err, nil
		}
		defer out.Close()

		// Write the body to file
		_, err = io.Copy(out, resp.Body)
		return err, nil
	}
}

// WriteYAML writes YAML data without preserving comments
func WriteYAML(filePath string, data interface{}) error {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(data); err != nil {
		return err
	}
	encoder.Close()

	return os.WriteFile(filePath, buf.Bytes(), 0644)
}
func IsWindows() bool {
	return os.PathSeparator == '\\'
}

func CapitalizeFirst(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
