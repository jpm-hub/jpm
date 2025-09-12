package install

import (
	"encoding/json"
	"errors"
	"fmt"
	COM "jpm/common"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

var jpmRepo string = "http://localhost/"
var tab string = "  "
var once sync.Once

// var mu sync.Mutex

type Repo struct {
	Alias          string `json:"alias"`
	Repo           string `json:"repo"`
	GroupID        string
	ArtVer         string
	ArtefactID     string
	DependencyType string
}

type Repositories struct {
	Repos []Repo `json:"repos"`
}

var scopesAccepted []string = []string{"compile", "runtime", ""}
var jpmScopes []string = []string{"test", "exec", ""}
var depsList []string = []string{}
var importList []string = []string{}
var currentWorkingRepo string
var currentOuterScope string
var cache map[string]pom = map[string]pom{}
var depMap map[string]string
var downloadInfo map[string][]string = map[string][]string{}
var failedInstalledList []string = []string{}
var latests []string = []string{}

func Install() {
	println()
	COM.FindPackageYML(true)
	os.MkdirAll(filepath.Join("jpm_dependencies", "tests"), 0755)
	homeDir := COM.HomeDir()
	if COM.IsWindows() {
		COM.RunCMD("copy "+filepath.Join(homeDir, "libs", "junit.jar")+" "+filepath.Join("jpm_dependencies", "tests", "junit.jar"), true)
	} else {
		COM.RunScript("cp "+filepath.Join(homeDir, "libs", "junit.jar")+" "+filepath.Join("jpm_dependencies", "tests", "junit.jar"), true)
	}
	if slices.Contains(os.Args, "-f") && !slices.Contains(os.Args, "-repo") {
		for i, arg := range os.Args {
			if arg == "-f" {
				os.Args = append(os.Args[:i], os.Args[i+1:]...)
				os.Remove(filepath.Join("jpm_dependencies", ".lock.json"))
				break
			}
		}
	}
	switch len(os.Args) {
	case 2:
		deps := COM.GetDependencies(false)
		aliases := findRepoAlias(deps)
		installFromYML(aliases, deps)
	case 4:
		if os.Args[2] == "-repo" {
			addNewRepo(os.Args[3])
			return
		}
		fallthrough
	default:
		installFromCLI()
	}
	if len(downloadInfo) != 0 {
		print("      Downloading [")
		for k, v := range downloadInfo {
			download(k, v[0], v[1], v[2], listAlreadyInstalledDeps())
			print("=")
		}
		println("]")
		for _, v := range failedInstalledList {
			println(v)
		}
	}

	dumpDependencies(depMap)
}
func QuickInstall() {
	COM.FindPackageYML(true)
	os.MkdirAll(filepath.Join("jpm_dependencies", "tests"), 0755)
	homeDir := COM.HomeDir()
	if COM.IsWindows() {
		COM.RunCMD("copy "+filepath.Join(homeDir, "libs", "junit.jar")+" "+filepath.Join("jpm_dependencies", "tests", "junit.jar"), true)
	} else {
		COM.RunScript("cp "+filepath.Join(homeDir, "libs", "junit.jar")+" "+filepath.Join("jpm_dependencies", "tests", "junit.jar"), true)
	}
	deps := COM.GetDependencies(false)
	aliases := findRepoAlias(deps)
	installFromYML(aliases, deps)
	if len(downloadInfo) != 0 {
		print("      Downloading [")
		for k, v := range downloadInfo {
			download(k, v[0], v[1], v[2], listAlreadyInstalledDeps())
			print("-")
		}
		println("]")
		for _, v := range failedInstalledList {
			println(v)
		}
	}
}
func installFromYML(aliases []string, deps []string) {
	// load in the lock.json
	loadDependencies()
	// raws first
	rawDeps := findAllRaws(deps)
	fromRAW(rawDeps)
	// jpm second
	jpmDeps := findAllJPM(deps, aliases)
	fromJPM(jpmDeps)
	// others last
	repoList := getRepoList()
	repoDeps := findAllRepoDeps(deps, repoList)
	fromRepo(repoDeps)
}

func fromJPM(deps []string) {
	for _, d := range deps {
		checkSlice := strings.Split(d, " ")
		if len(checkSlice) > 2 {
			println(tab + "JPM Dependency " + d + " Does not exist")
			continue
		}
		println("\033[32m  --- JPM: Resolving " + d + "\033[0m")
		dSlices := strings.Split(d, " ")
		dependencyType := ""
		if len(dSlices) > 1 && checkValidScope(dSlices[1]) {
			dependencyType = dSlices[1]
		}
		depAndVersion := strings.Split(dSlices[0], ":")
		dep := dSlices[0]
		version := ""
		if len(depAndVersion) > 1 {
			if depAndVersion[1] == "latest" {
				// figure out version for this jpm dependency
			}
			version = depAndVersion[1]
		} else {
			// figure out version for this jpm dependency
			// modify the yaml at this point
			COM.ReplaceDependency(d, fmt.Sprintf("%s:%s %s", dep, version, dependencyType))
		}
		filename := dep + "-" + version + ".jar"
		url := jpmRepo + dep + "/" + version + "/" + filename
		downloadInfo[url] = []string{filename, "jpm", dependencyType}
	}
}
func fromRepo(dependenciesWithRepo []Repo) {
	for _, dr := range dependenciesWithRepo {
		if !checkValidScope(dr.DependencyType) {
			currentOuterScope = ""
		} else {
			currentOuterScope = dr.DependencyType
		}
		currentWorkingRepo = dr.Repo
		saveAllRepoSubDependencies(&dr)
	}
	if len(depsList) > 0 {
		dr := dependenciesWithRepo[0]
		addRepoSubDependenciesToDownloadList(dr)
	}
}
func addRepoSubDependenciesToDownloadList(dr Repo) {
	depMap = resolveDependecy()
	for k, v := range depMap {
		gas := strings.Split(k, "|")
		groupID := gas[0]
		artefactID := gas[1]
		version := v
		scope := ""
		if len(gas) == 3 && len(gas[2]) > 1 {
			scope = gas[2]
		}
		url := dr.Repo + strings.ReplaceAll(groupID, ".", "/") + "/" + artefactID + "/" + version + "/" + artefactID + "-" + version + ".jar"
		filename := artefactID + "-" + version + ".jar"

		downloadInfo[url] = []string{filename, dr.Alias, "" + scope + "|"}
	}

	depsList = []string{}
}
func saveAllRepoSubDependencies(dr *Repo) error {
	version := dr.ArtVer
	var err error
	if dr.ArtVer == "" {
		version, err = figureOutLastest(dr.GroupID, dr.ArtefactID)
		if err != nil {
			println("\033[31m  --- " + strings.ToUpper(dr.Alias) + ": Resolving " + dr.ArtefactID + " ! " + "Unable to get latest version: " + err.Error() + " \033[0m\n")
			return errors.New("not nil")
		}
		dr.ArtVer = version
		// modify the yaml at this point
		COM.ReplaceDependency(fmt.Sprintf("%s %s %s %s", dr.Alias, dr.GroupID, dr.ArtefactID, dr.DependencyType), fmt.Sprintf("%s %s %s:%s %s", dr.Alias, dr.GroupID, dr.ArtefactID, dr.ArtVer, dr.DependencyType))
	}
	if dr.ArtVer == "latest" {
		version, err = figureOutLastest(dr.GroupID, dr.ArtefactID)
		if err != nil {
			println("\033[31m  --- " + strings.ToUpper(dr.Alias) + ": Resolving " + dr.ArtefactID + " ! " + "Unable to get latest version: " + err.Error() + " \033[0m\n")
			return errors.New("not nil")
		}
		dr.ArtVer = "latest"
	}
	println("\033[32m  --- " + strings.ToUpper(dr.Alias) + "\033[0m: Resolving " + dr.ArtefactID + ":" + version)
	print("      Resolving   [")
	downloadDepsRepo(dr.Repo, dr.GroupID, dr.ArtefactID, version, false)
	figureOutAllLatestAndDownload()
	println("]")
	return nil
}
func fromRAW(deps []string) {
	for _, d := range deps {
		checkSlice := strings.Split(d, " ")
		if len(checkSlice) > 4 || len(checkSlice) < 2 {
			println(tab + "RAW Dependency " + d + " Does not exist")
			continue
		}
		println("  --- RAW: Resolving " + d)
		dSlices := strings.Split(d, " ")
		dependencyType := ""
		url := dSlices[1]
		if len(dSlices) == 3 && !strings.Contains(d, " -x ") && checkValidScope(dSlices[2]) {
			dependencyType = dSlices[2]
		} else if len(dSlices) == 3 && strings.Contains(d, " -x ") {
			url = dSlices[2]
			dependencyType = "jpm_dependencies x"
		} else if len(dSlices) == 4 && checkValidScope(dSlices[3]) {
			url = dSlices[2]
			dependencyType = dSlices[3] + " x"
		}
		urlSlice := strings.Split(url, "/")
		filename := urlSlice[len(urlSlice)-1]
		downloadInfo[url] = []string{filename, "raw", dependencyType}
	}
}
func checkValidScope(scope string) bool {
	if !slices.Contains(jpmScopes, scope) {
		println()
		println(tab + " scope " + scope + " is unknown scope")
		println(tab + " Will be considered as normal dependency")
		println()
		return false
	}
	return true
}
func download(url string, filename string, alias string, scope string, depsInstalled map[string][]string) {
	sc := strings.Split(strings.Trim(scope, "|"), "|")[0]
	extract := false
	if len(strings.Split(sc, " ")) == 2 {
		sc = strings.Split(sc, " ")[0]
		extract = true
	}
	if sc == "exec" {
		once.Do(func() {
			os.MkdirAll(filepath.Join("jpm_dependencies", "execs"), 0755)
		})
	}
	switch sc {
	case "test":
		v, ok := depsInstalled["test"]
		if ok && slices.Contains(v, filename) {
			return
		}
		if err, _ := COM.DownloadFile(url, filepath.Join("jpm_dependencies", "tests"), filename, false, false); err != nil {
			if err.Error() == "404 Not Found" {
				return
			}
			failedInstalledList = append(failedInstalledList, tab+"Failed to correctly install : "+filename+" ERR:"+err.Error())
			return
		}
		if extract {
			COM.Extract(filepath.Join("jpm_dependencies", "tests"), filename)
			COM.CleanupExtract(filepath.Join("jpm_dependencies", "tests"), filename)
		}
	case "exec":
		v, ok := depsInstalled["exec"]
		if ok && slices.Contains(v, filename) {
			return
		}
		if err, _ := COM.DownloadFile(url, filepath.Join("jpm_dependencies", "execs"), filename, false, false); err != nil {
			if err.Error() == "404 Not Found" {
				return
			}
			failedInstalledList = append(failedInstalledList, tab+"Failed to correctly install : "+filename+" ERR:"+err.Error())
			return
		}
		if extract {
			COM.Extract(filepath.Join("jpm_dependencies", "execs"), filename)
			COM.CleanupExtract(filepath.Join("jpm_dependencies", "execs"), filename)
		}
	default:
		v, ok := depsInstalled["jpm_dependencies"]
		if ok && slices.Contains(v, filename) {
			return
		}
		if err, _ := COM.DownloadFile(url, "jpm_dependencies", filename, false, false); err != nil {
			if err.Error() == "404 Not Found" {
				return
			}
			failedInstalledList = append(failedInstalledList, tab+"Failed to correctly install : "+filename+" ERR:"+err.Error())
			return
		}
		if extract {
			COM.Extract("jpm_dependencies", filename)
			COM.CleanupExtract("jpm_dependencies", filename)
		}
	}

}
func getRepoList() Repositories {
	repoSection := COM.GetSection("repos", false)
	repos := Repositories{
		Repos: []Repo{},
	}
	if repoSection == nil {
		return repos
	}
	repoSectionListMap := repoSection.([]any)
	for _, r := range repoSectionListMap {
		repoSectionMap := r.(map[string]string)
		//for now jpm will be local
		for k, v := range repoSectionMap {
			url := v
			if !strings.HasSuffix(url, "/") {
				url = url + "/"
			}
			repos.Repos = append(repos.Repos, Repo{Alias: strings.ToLower(k), Repo: url})
		}
	}
	return repos
}

func findRepoAlias(dependencyList []string) []string {
	aliases := []string{}
	for _, v := range dependencyList {
		a := strings.Split(v, " ")[0]
		if slices.Contains(aliases, a) {
			continue
		}
		aliases = append(aliases, strings.ToLower(a))
	}
	return aliases
}
func findAllRepoDeps(deps []string, repoList Repositories) []Repo {
	repos := []Repo{}
	for _, v := range deps {
		found := false
		for _, repoFromYaml := range repoList.Repos {
			if len(repoFromYaml.Alias) < 1 || COM.NormalizeSpaces(repoFromYaml.Alias) == "" || repoFromYaml.Alias == "" {
				continue
			}
			v = COM.NormalizeSpaces(v)
			if strings.HasPrefix(v, repoFromYaml.Alias) {
				found = true
				found2 := false
				if !found2 {
					r, err := disectRepoDepString(v, repoFromYaml.Repo, repoFromYaml.Alias)
					if err == nil {
						repos = append(repos, r)
					}
				}
			}
		}
		if strings.HasPrefix(COM.NormalizeSpaces(strings.ToLower(v)), "raw ") {
			continue
		}
		if !found {
			println("\033[31m" + tab + "The repository for this dependency was not found: " + v + "\033[0m\n")
		}

	}

	return repos
}
func disectRepoDepString(depString string, repoURL string, alias string) (Repo, error) {
	dSlice := strings.Split(depString, " ")
	if len(dSlice) < 3 {
		println("\033[31m" + tab + "The dependency : " + depString + " is ambigious\033[0m")
		return Repo{}, errors.New("The dependency : " + depString + " is ambigious")
	}
	artver := strings.Split(dSlice[2], ":")
	t := ""
	if len(dSlice) == 3 {
		t = ""
	} else if len(dSlice) == 4 {
		t = dSlice[3]
	} else {
		println("\033[31m" + tab + "The dependency : " + depString + " is ambigious\033[0m")
		return Repo{}, errors.New("The dependency : " + depString + " is ambigious")
	}
	version := ""
	if len(artver) == 2 {
		version = artver[1]
	}
	return Repo{
		Repo:           repoURL,
		Alias:          alias,
		GroupID:        dSlice[1],
		ArtefactID:     artver[0],
		ArtVer:         COM.NormalizeSpaces(version),
		DependencyType: t,
	}, nil
}
func findAllJPM(unfiltterenDeps []string, aliases []string) []string {
	deps := []string{}
	for _, v := range unfiltterenDeps {
		v = COM.NormalizeSpaces(v)
		if !slices.Contains(aliases, strings.Split(v, " ")[0]) {
			// Skip if we already have this dependency
			found := slices.Contains(deps, v)
			if !found {
				deps = append(deps, v)
			}

		}

	}
	return deps
}
func findAllRaws(unfiltterenDeps []string) []string {
	deps := []string{}
	for _, v := range unfiltterenDeps {
		if strings.HasPrefix(COM.NormalizeSpaces(strings.ToLower(v)), "raw") {
			// Skip if we already have this dependency
			found := slices.Contains(deps, v)
			if !found {
				deps = append(deps, v)
			}
		}
	}
	return deps
}
func listAlreadyInstalledDeps() map[string][]string {
	dir := "jpm_dependencies"
	entries, err := os.ReadDir(dir)
	if err != nil {
		println("Could not fetch list of already installed jars")
		return map[string][]string{}
	}
	list := []string{}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			continue
		}
		list = append(list, name)
	}
	maped := make(map[string][]string, 3)
	maped["jpm_dependencies"] = list
	testList := []string{}
	dir = "jpm_dependencies/tests"
	entries, err = os.ReadDir(dir)
	if err != nil {
		println("Could not fetch list of already installed test jars")
		return maped
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			continue
		}
		testList = append(testList, name)
	}
	maped["test"] = testList
	execList := []string{}
	dir = "jpm_dependencies/execs"
	entries, err = os.ReadDir(dir)
	if err != nil {
		return maped
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			continue
		}
		execList = append(execList, name)
	}
	maped["exec"] = execList
	return maped
}
func resolveDependecy() map[string]string {
	depMap := make(map[string]string, len(depsList)/2)
	for i := 0; i < len(depsList); i += 2 {
		if depMap[depsList[i]] == "" {
			depMap[depsList[i]] = depsList[i+1]
		} else if depsList[i+1] == "" {
			continue
		} else if depMap[depsList[i]] < depsList[i+1] {
			depMap[depsList[i]] = depsList[i+1]
		}
	}
	for key, value := range depMap {
		// try finding it in imports
		if value == "" {
			for i, v := range importList {
				if v == key {
					depMap[key] = importList[i+1]
				}
			}
		}
	}
	for k := range depMap {
		if strings.HasSuffix(k, "|test") {
			d := strings.TrimSuffix(k, "test")
			if _, ok := depMap[d]; ok {
				depMap[k] = ""
			}
		}
	}
	maps.DeleteFunc(depMap, func(k string, v string) bool { return v == "" })
	return depMap
}
func loadDependencies() {
	file, err := os.Open(filepath.Join("jpm_dependencies", ".lock.json"))
	if err != nil {
		// If file does not exist, just return
		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	lockMap := make(map[string]string)
	if err := decoder.Decode(&lockMap); err != nil {
		// If decode fails, just return
		return
	}

	for k, v := range lockMap {
		depsList = append(depsList, k, v)
	}
}

func dumpDependencies(depMap map[string]string) {
	lockFilePath := filepath.Join("jpm_dependencies", ".lock.json")

	// Remove the file if it already exists to ensure we override it
	if _, err := os.Stat(lockFilePath); err == nil {
		if err := os.Remove(lockFilePath); err != nil {
			fmt.Println("Error removing existing lock.json:", err)
			return
		}
	}
	jars := []string{}
	for key, v := range depMap {
		jar := ""
		if strings.HasSuffix(key, "|test") {
			jar = filepath.Join("tests")
		}
		value := strings.TrimSuffix(key, "|test")
		valueS := strings.Split(strings.TrimSuffix(value, "|"), "|")
		value = valueS[len(valueS)-1]
		jar = filepath.Join(jar, value+"-"+v+".jar")
		jars = append(jars, jar)
		depMap[key] = v
	}

	file, err := os.Create(lockFilePath)
	if err != nil {
		fmt.Println("Error creating .lock.json:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(depMap); err != nil {
		fmt.Println("Error encoding .lock.json:", err)
	}

	//clean up here
	files, err := os.ReadDir("jpm_dependencies")
	if err != nil {
		fmt.Println("Error reading jpm_dependencies directory:", err)
		return
	}
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".jar") {
			if !slices.Contains(jars, file.Name()) {
				os.Remove(filepath.Join("jpm_dependencies", file.Name()))
			}
		}
	}
	files, err = os.ReadDir(filepath.Join("jpm_dependencies", "tests"))
	if err != nil {
		fmt.Println("Error reading jpm_dependencies/tests directory:", err)
		return
	}
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".jar") {
			if !slices.Contains(jars, filepath.Join("tests", file.Name())) {
				if file.Name() == "junit.jar" {
					continue
				}
				os.Remove(filepath.Join("jpm_dependencies", "tests", file.Name()))
			}
		}
	}
}
