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
	"runtime"
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
var parsed []string = []string{}

type Dependencies struct {
	Classified   bool                         `json:"classified"`
	Dependencies []string                     `json:"dependencies"`
	JPM          map[string]string            `json:"JPM"`
	Repos        map[string]map[string]string `json:"repos"`
	Scripts      map[string]string            `json:"scripts,omitempty"`
}
type PackageYAMLSimple struct {
	Main          string              `yaml:"main,omitempty"`
	Package       string              `yaml:"package,omitempty"`
	Packages      []string            `yaml:"packages,omitempty"`
	Src           string              `yaml:"src,omitempty"`
	Version       string              `yaml:"version,omitempty"`
	Description   string              `yaml:"description,omitempty"`
	Language      string              `yaml:"language,omitempty"`
	Env           string              `yaml:"env,omitempty"`
	Classified    bool                `yaml:"classified,omitempty"`
	Modular       bool                `yaml:"modular,omitempty"`
	Scripts       map[string]string   `yaml:"scripts,omitempty"`
	Dependencies  []string            `yaml:"dependencies,omitempty"`
	Classifiers   map[string]string   `yaml:"classifiers,omitempty"`
	Repos         []map[string]string `yaml:"repos,omitempty"`
	Args          map[string]string   `yaml:"args,omitempty"`
	Excludes      []string            `yaml:"excludes,omitempty"`
	OtherSections map[string]any      `yaml:",inline,omitempty"`
}

// PackageYAML represents the structure of package.yml with ordered fields
type PackageYAML struct {
	Main          string         `yaml:"main,omitempty"`
	Package       string         `yaml:"package,omitempty"`
	Packages      []string       `yaml:"packages,omitempty"`
	Src           string         `yaml:"src,omitempty"`
	Version       string         `yaml:"version,omitempty"`
	Description   string         `yaml:"description,omitempty"`
	Language      string         `yaml:"language,omitempty"`
	Env           string         `yaml:"env,omitempty"`
	Classified    bool           `yaml:"classified,omitempty"`
	Modular       bool           `yaml:"modular,omitempty"`
	Scripts       *OrderedMap    `yaml:"scripts,omitempty"`
	Dependencies  []string       `yaml:"dependencies"`
	Classifiers   *OrderedMap    `yaml:"classifiers,omitempty"`
	Repos         []*OrderedMap  `yaml:"repos,omitempty"`
	Args          *OrderedMap    `yaml:"args,omitempty"`
	Excludes      []string       `yaml:"excludes,omitempty"`
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
		Verbose bool `json:"Verbose"`
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
	kotlinc := filepath.Join(HomeDir(), "kotlinc", "bin", "kotlinc.bat")
	if _, err := os.Stat(kotlinc); !os.IsNotExist(err) {
		return kotlinc
	}
	return "kotlinc"
}

func CheckDeps(dep string) {
	deps := StripVersionInfo(NormalizeDependencies(GetDependencies(true))...)
	seen := make(map[string]bool)
	seen[StripVersionInfo(NormalizeSpaces(dep))[0]] = true
	for _, d := range deps {
		if seen[NormalizeSpaces(d)] {
			println("package.yml:", d, "would be duplicated in dependencies")
			println("Aborting...")
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
				k = strings.ToLower(k)
				if strings.HasPrefix(k, "-") || slices.Contains(alreadyThere, k) || k == "local" || k == "raw" || k == "jpm" || v == "" || k == "_" {
					println("syntax error with: ", k, v)
					os.Exit(1)
				}
				if k != "username" && k != "password" && k != "type" {
					alreadyThere = append(alreadyThere, k)
				}
			}
		}
	}
}
func FindPackageYML(fatal bool) (string, string) {
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
	if fatal {
		os.Exit(404)
	}
	return "", ""
}
func StripVersionInfo(deps ...string) []string {
	result := []string{}
	for _, dep := range NormalizeDependencies(deps) {
		if strings.HasPrefix(dep, "raw ") || strings.HasPrefix(dep, "local ") {
			result = append(result, strings.TrimSpace(dep))
			continue
		}
		indexOflastcolon := strings.Index(dep, ":")
		if indexOflastcolon == -1 {
			result = append(result, strings.TrimSpace(dep))
			continue
		}
		indexOfSpaceAfterColon := strings.Index(dep[indexOflastcolon:], " ")
		if indexOfSpaceAfterColon != -1 {
			// string removing content between last colon and space after it
			dep = dep[:indexOflastcolon] + dep[indexOflastcolon+indexOfSpaceAfterColon:]
		} else {
			// string removing content after last colon
			dep = dep[:indexOflastcolon]
		}
		result = append(result, strings.TrimSpace(dep))
	}
	return result
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
	replaced := false
	for i, depStr := range pkgYAML.Dependencies {
		if strings.Count(depStr, ":") > 1 {
			depStr = strings.Replace(depStr, ":", " ", 1)
		}
		depStr = StripVersionInfo(NormalizeSpaces(depStr))[0]
		oldDepString = StripVersionInfo(NormalizeSpaces(oldDepString))[0]
		if depStr != oldDepString {
			continue
		}
		replaced = true
		newDepString = NormalizeSpaces(newDepString)
		pkgYAML.Dependencies[i] = newDepString
		packageYML.Dependencies[i] = newDepString
		break
	}
	if !replaced {
		AddToSection("dependencies", newDepString)
		return
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

	if sectionValue == nil {
		return
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
			packageYML.Dependencies = make([]string, 0)
		}
		newDependency := sectionValue.(string)
		CheckDeps(newDependency)
		if !slices.Contains(pkgYAML.Dependencies, newDependency) {
			pkgYAML.Dependencies = append(pkgYAML.Dependencies, newDependency)
			packageYML.Dependencies = append(packageYML.Dependencies, newDependency)
		}
	case "repos":
		if pkgYAML.Repos == nil {
			pkgYAML.Repos = []*OrderedMap{}
			packageYML.Repos = []map[string]string{}
		}

		if repoMap, ok := sectionValue.(map[string]string); ok {
			newRepos := &OrderedMap{
				Values: map[string]any{},
				Order:  []string{},
			}
			// we already check if it exists when calling -repo
			for k, v := range repoMap {
				newRepos.Values[k] = v
				newRepos.Order = append(newRepos.Order, k)
			}
			pkgYAML.Repos = append(pkgYAML.Repos, newRepos)
			// Convert map[string]any to map[string]string for packageYML.Repos
			converted := make(map[string]string)
			for k, v := range newRepos.Values {
				if strVal, ok := v.(string); ok {
					converted[k] = strVal
				}
			}
			packageYML.Repos = append(packageYML.Repos, converted)
		}
	case "classifiers":
		if pkgYAML.Classifiers == nil {
			pkgYAML.Classifiers = &OrderedMap{
				Values: make(map[string]any),
				Order:  make([]string, 0),
			}
			packageYML.Classifiers = make(map[string]string)
		}
		if classifierMap, ok := sectionValue.(map[string]string); ok {
			for k, v := range classifierMap {
				pkgYAML.Classifiers.Values[k] = v
				packageYML.Classifiers[k] = v
				pkgYAML.Classifiers.Order = append(pkgYAML.Classifiers.Order, k)
			}
		}
	case "excludes":
		if pkgYAML.Excludes == nil {
			pkgYAML.Excludes = []string{}
			packageYML.Excludes = []string{}
		}
		newExclude := sectionValue.(string)
		if !slices.Contains(pkgYAML.Excludes, newExclude) {
			pkgYAML.Excludes = append(pkgYAML.Excludes, newExclude)
			packageYML.Excludes = append(packageYML.Excludes, newExclude)
		}
	}
	if err := WriteYAML(g_yamlPath, pkgYAML); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func GetSection(section string, isFatal bool) any {
	if packageYML.Package == "" && g_yamlPath != "" {
		println("package section of package.yml cannot be empty")
		os.Exit(1)
	}
	switch section {
	case "main":
		return ParseENV(packageYML.Main)
	case "language":
		lang := ParseENV(packageYML.Language)
		if lang == "" {
			return "java"
		}
		return lang
	case "version":
		return ParseENV(packageYML.Version)
	case "description":
		return ParseENV(packageYML.Description)
	case "env":
		return ParseENV(packageYML.Env)
	case "package":
		return ParseENV(packageYML.Package)
	case "packages":
		packages := []string{}
		for _, v := range packageYML.Packages {
			packages = append(packages, ParseENV(v))
		}
		return packages
	case "src":
		return strings.TrimSpace(ParseENV(packageYML.Src))
	case "classified":
		return packageYML.Classified
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
	case "classifiers":
		classes := map[string]string{}
		for key, v := range packageYML.Classifiers {
			classes[key] = ParseENV(v)
		}
		return classes
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
	case "modular":
		return packageYML.Modular
	case "excludes":
		excludes := []string{}
		for _, v := range packageYML.Excludes {
			excludes = append(excludes, ParseENV(v))
		}
		return excludes
	default:
		if val, ok := packageYML.OtherSections[section]; ok {
			other, ok := val.(string)
			if !ok {
				if isFatal {
					println("Error: section", section, "is not a string")
					os.Exit(1)
				}
				return val
			}
			if other == "" {
				if isFatal {
					println("Error: section", section, "is empty")
					os.Exit(1)
				}
				return ""
			}
			return ParseENV(other)
		} else {
			if isFatal {
				println("Error: section", section, "not found")
				os.Exit(1)
			}
			return nil
		}

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
				env["env."+key] = val
			}
		}
	}
	env["jpm.OS"] = runtime.GOOS
	env["jpm.ARCH"] = runtime.GOARCH
	env["jpm.OS-ARCH"] = runtime.GOOS + "-" + runtime.GOARCH
	env["main"] = ParseENV(packageYML.Main)
	env["package"] = ParseENV(packageYML.Package)
	env["src"] = ParseENV(packageYML.Src)
	env["version"] = ParseENV(packageYML.Version)
	env["language"] = ParseENV(packageYML.Language)
	for k, v := range packageYML.OtherSections {
		if str, ok := v.(string); ok {
			env[k] = ParseENV(str)
		} else {
			env[k] = fmt.Sprintf("%v", v)
		}
	}
}
func ParseENV(str string) string {
	result := str
	if _, ok := env[":"]; !ok {
		separator := ":"
		if IsWindows() {
			separator = ";"
		}
		env[":"] = separator
	}
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
		if !ok && varName != "" {
			if val, ok = os.LookupEnv(strings.TrimPrefix(varName, "ENV.")); !ok {
				found := false
				if slices.Contains(parsed, varName) {
					fmt.Println("\nError :", "\033[31m circular reference detected with", str, " : ", parsed, "\033[0m")
					os.Exit(1)
					return ""
				}
				parsed = append(parsed, varName)
				for k, v := range packageYML.OtherSections {
					if k == varName {
						env[k] = ParseENV(fmt.Sprintf("%v", v))
						val = env[k]
						found = true
						break
					}
				}
				if !found {
					println("\nError :", "\033[31m{{", varName, "}} was not found anywhere\033[0m")
					// If not found, remove the unmatched {{KEY}} to avoid infinite loop
					result = result[:startIdx] + result[endIdx+2:]
					continue
				}
			}
		}
		result = result[:startIdx] + val + result[endIdx+2:]
	}

	return result
}
func ParseEnvVars(prefix string, quotes bool) string {
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
			if quotes {
				lines = append(lines, prefix+l[0]+"='"+l[1]+"'")
			} else {
				lines = append(lines, prefix+l[0]+"="+l[1])
			}
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
	if depListRawNil == nil || len(depListRawNil.([]string)) == 0 {
		return []string{}
	}
	os.MkdirAll(filepath.Join("jpm_dependencies", "tests"), 0755)
	depListRaw := depListRawNil.([]string)
	var deps []string = []string{}
	for _, item := range depListRaw {
		deps = append(deps, strings.TrimSpace(item))

	}
	return NormalizeDependencies(deps)
}

func NormalizeDependencies(dep []string) []string {
	for i, v := range dep {
		if strings.HasPrefix(v, "raw ") || strings.HasPrefix(v, "local ") {
			continue
		}
		if strings.Count(v, ":") > 1 {
			dep[i] = strings.Replace(dep[i], ":", " ", 1)
		}
		dep[i] = NormalizeSpaces(dep[i])
	}
	return dep
}
func RunPS(script string, showStdOut bool) error {
	if !IsWindows() {
		println("PS is only available on windows")
		return fmt.Errorf("PS is only available on windows")
	}
	var cmd *exec.Cmd
	if tmpFile, err := os.CreateTemp("", "jpm_script_*.ps1"); err == nil {
		tmpFile.WriteString(script)
		if Verbose {
			println("\033[33m-(Verbose)=> " + script + "\033[0m")
			showStdOut = true
		}
		tmpFile.Close()
		cmd = exec.Command("powershell", "-File", tmpFile.Name())
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
				return fmt.Errorf("[JPM] signal: %v", sig)
			}
		}
		return fmt.Errorf("failed to trap signals")
	}
	return fmt.Errorf("failed to run")
}
func RunCMD(script string, showStdOut bool) error {
	if !IsWindows() {
		println("CMD is only available on windows")
		return fmt.Errorf("CMD is only available on windows")
	}
	var cmd *exec.Cmd
	if tmpFile, err := os.CreateTemp("", "jpm_script_*.cmd"); err == nil {
		tmpFile.WriteString("@echo off\nchcp 65001 > NUL 2>&1\n" + script)
		if Verbose {
			println("\033[33m-(Verbose)=>  " + script + "\033[0m")
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
				return fmt.Errorf("[JPM] signal: %v", sig)
			}
		}
		return fmt.Errorf("failed to trap signals")
	}
	return fmt.Errorf("failed to run")
}
func RunScript(script string, showStdOut bool) error {
	var cmd *exec.Cmd
	if IsWindows() {
		cmd = exec.Command("C:\\Program Files\\Git\\bin\\bash.exe", "-c", script)
	} else {
		cmd = exec.Command("sh", "-c", script)
	}
	if dir, err := os.Getwd(); err == nil {
		cmd.Dir = dir
	}
	if Verbose {
		println("\033[33m-(Verbose)=> " + script + "\033[0m")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
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
		return "."
	}
	appDir := ".jpm"
	// Create the directory if it doesn't exist
	dirPath := filepath.Join(homeDir, appDir)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			panic(err)
		}
	}
	return dirPath
}

func HomeDirUnix() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	appDir := ".jpm"
	// Create the directory if it doesn't exist
	dirPath := filepath.Join(homeDir, appDir)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			println("failed to create home directory:", err.Error())
			os.Exit(1)
		}
	}
	dirPath = strings.ReplaceAll(dirPath, ":", "")
	prts := strings.Split(dirPath, "\\")
	return "/" + strings.ToLower(prts[0]) + "/" + strings.Join(prts[1:], "/")
}

func SrcDir() string {
	valsrc := GetSection("src", false)
	if str := valsrc.(string); str != "" {
		return strings.TrimSuffix(str, "/")
	} else {
		valpkg := GetSection("package", true)
		if str = valpkg.(string); str != "" {
			_, err := os.Stat(strings.ReplaceAll(str, "-", "_"))
			if err == nil {
				return strings.ReplaceAll(str, "-", "_")
			}
			return "."
		}
		return "."
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
	err := os.Remove(filepath.Join(dirname, filename))
	if err != nil {
		//println("failed to remove", filename, ":", err.Error())
	}
}
func DownloadFile(url string, dirpath string, filename string, override bool, returnContent bool) (error, []byte) {
	filePath := filepath.Join(dirpath, filename)
	// Get the data first to check response
	resp, err := http.Get(url)
	if err != nil {
		if Verbose {
			println("\033[31m" + url + "\033[0m")
		}
		return err, nil
	}
	defer resp.Body.Close()

	if _, err := os.Stat(filePath); err == nil {
		if override {
			os.Remove(filePath)
		} else {
			if Verbose {
				println(filename, "already exists, skipping download")
			}
			return errors.New("file already exist"), nil
		}
	}

	// Check server response before creating file
	if resp.StatusCode <= http.StatusOK && resp.StatusCode >= 300 {
		if Verbose {
			println("\033[31m" + url + "\033[0m")
		}
		return fmt.Errorf("%s", resp.Status), nil
	}

	if returnContent {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			if Verbose {
				println("\033[31m" + url + "\033[0m")
			}
			return err, nil
		}
		return nil, bodyBytes
	} else {
		// Create the file only after successful HTTP response
		out, err := os.Create(filePath)
		if err != nil {
			if Verbose {
				println("\033[31m" + url + "\033[0m")
			}
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
	sS := strings.Split(s, "-")
	for i, v := range sS {
		runes := []rune(v)
		runes[0] = unicode.ToUpper(runes[0])
		sS[i] = string(runes)
	}
	return strings.Join(sS, "")
}

func Sanitize(s string, canBeEmpty bool) string {
	var b strings.Builder
	if len(s) == 0 && !canBeEmpty {
		println("cannot be empty")
		os.Exit(1)
	}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			println("contains invalid character in :", s)
			os.Exit(1)
		}
	}
	return b.String()
}

func CopyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func CopyToDependencies(lang string) {
	homeDir := HomeDir()
	os.MkdirAll(filepath.Join("jpm_dependencies", "tests"), 0755)
	CopyFile(filepath.Join(homeDir, "libs", "junit.jar"), filepath.Join("jpm_dependencies", "tests", "junit.jar"))
	if IsWindows() {
		homeDir = filepath.Join(homeDir, "kotlinc", "lib")
	} else {
		homeDir = filepath.Join(homeDir, "libs")
	}
	if strings.Contains(lang, "kotlin") {
		if err := CopyFile(filepath.Join(homeDir, "kotlin-test.jar"), filepath.Join("jpm_dependencies", "tests", "kotlin-test.jar")); err != nil {
			fmt.Printf("Error copying kotlin-test.jar: %v\n", err)
		}
		if err := CopyFile(filepath.Join(homeDir, "annotations-13.0.jar"), filepath.Join("jpm_dependencies", "annotations-13.0.jar")); err != nil {
			fmt.Printf("Error copying annotations-13.0.jar: %v\n", err)
		}
		if err := CopyFile(filepath.Join(homeDir, "kotlin-stdlib.jar"), filepath.Join("jpm_dependencies", "kotlin-stdlib.jar")); err != nil {
			fmt.Printf("Error copying kotlin-stdlib.jar: %v\n", err)
		}
		if err := CopyFile(filepath.Join(homeDir, "kotlin-reflect.jar"), filepath.Join("jpm_dependencies", "kotlin-reflect.jar")); err != nil {
			fmt.Printf("Error copying kotlin-reflect.jar: %v\n", err)
		}
	}
}

func LinkToDependencies(lang string) {
	homeDir := HomeDir()
	os.MkdirAll(filepath.Join("jpm_dependencies", "tests"), 0755)
	junitPath := filepath.Join("jpm_dependencies", "tests", "junit.jar")
	if _, err := os.Stat(junitPath); os.IsNotExist(err) {
		os.Link(filepath.Join(homeDir, "libs", "junit.jar"), junitPath)
	}
	if IsWindows() {
		homeDir = filepath.Join(homeDir, "kotlinc", "lib")
	} else {
		homeDir = filepath.Join(homeDir, "libs")
	}
	if strings.Contains(lang, "kotlin") {
		kotlinFiles := []string{"kotlin-test.jar", "annotations-13.0.jar", "kotlin-stdlib.jar", "kotlin-reflect.jar"}
		for _, file := range kotlinFiles {
			var destPath string
			if file == "kotlin-test.jar" {
				destPath = filepath.Join("jpm_dependencies", "tests", file)
			} else {
				destPath = filepath.Join("jpm_dependencies", file)
			}
			
			if _, err := os.Stat(destPath); os.IsNotExist(err) {
				srcPath := filepath.Join(homeDir, file)
				if err := os.Link(srcPath, destPath); err != nil {
					fmt.Printf("Error linking %s: %v\n", file, err)
				}
			}
		}
	}
}
func Ping(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
