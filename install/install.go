package install

import (
	"encoding/json"
	"fmt"
	COM "jpm/common"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

var jpmRepoUrl string = COM.JPM_REPO_API + "packages/"
var tab string = "  "
var once sync.Once

// var mu sync.Mutex

type Repo struct {
	Alias      string
	Repo       string
	Username   string
	Password   string
	GroupID    string
	ArtVer     string
	ArtefactID string
	Scope      string
}

type Repositories struct {
	Repos []Repo `json:"repos"`
}

var scopesAccepted []string = []string{"compile", "runtime", ""}
var jpmScopes []string = []string{"test", "exec", ""}
var depsList map[string][]string = map[string][]string{}
var importList []string = []string{}
var currentWorkingRepo string
var currentOuterScope string
var currentLanguage string = "java"
var cache map[string]pom = map[string]pom{}
var g_lockDeps COM.Dependencies = COM.Dependencies{}
var downloadInfo map[string][]string = map[string][]string{}
var failedInstalledList []string = []string{}
var latests []string = []string{}
var finishMessages []string = []string{}
var excludes []string = []string{}

func Install() {
	println()
	force := false
	COM.FindPackageYML(true)
	for i := 0; i < len(os.Args); i++ {
		if i == 0 || i == 1 {
			continue
		}

		if os.Args[i] == "-update" {
			os.Args = []string{os.Args[0], "install", "-update"}
			break
		}
		if os.Args[i] == "-v" {
			COM.Verbose = true
			os.Args = append(os.Args[:i], os.Args[i+1:]...)
			i--
			continue
		}
		if os.Args[i] == "-no" {
			COM.AddToSection("excludes", os.Args[i+1])
			os.Args = append(os.Args[:i], os.Args[i+2:]...)
			i--
			continue
		}
		if os.Args[i] == "-with" {
			COM.AddToSection("classifiers", determineCLIClassifier(os.Args[i+1]))
			os.Args = append(os.Args[:i], os.Args[i+2:]...)
			i--
			continue
		}
		if os.Args[i] == "-repo" {
			addNewRepo(os.Args[i+1])
			os.Args = append(os.Args[:i], os.Args[i+2:]...)
			i--
			continue
		}
		if os.Args[i] == "-f" {
			if force {
				println(" -f is used too many times!")
			}
			if !force && !COM.Ping("https://github.com") {
				println("\tCould not reach jpm repo at " + jpmRepoUrl)
				println("\tPlease check your internet connection or try again later")
				os.Exit(1)
			}
			os.Args = append(os.Args[:i], os.Args[i+1:]...)
			force = true
			i--
			continue
		}
	}
	COM.CopyToDependencies(COM.GetSection("language", true).(string))
	excludes = COM.GetSection("excludes", false).([]string)
	currentLanguage = COM.GetSection("language", false).(string)
	switch len(os.Args) {
	case 2:
		if force {
			os.Remove(filepath.Join("jpm_dependencies", "lock.json"))
		}
		// load in the lock.json
		lockDeps := loadLockDependencies()
		yamlDeps := COM.GetDependencies(false)
		deps := []string{}
		for _, dep := range yamlDeps {
			if !slices.Contains(lockDeps, dep) {
				deps = append(deps, dep)
			}
		}
		aliases := findExistingAliases()
		installFromYML(aliases, deps, true)
	case 3:
		if os.Args[2] == "-update" {
			os.Remove(filepath.Join("jpm_dependencies", "lock.json"))
			loadLockDependencies()
			deps := COM.GetDependencies(false)
			deps = COM.StripVersionInfo(deps...)
			aliases := findExistingAliases()
			installFromYML(aliases, deps, true)
			break
		}
		fallthrough
	default:
		if force {
			println("\033[31mcannot force install\033[0m")
		}
		aliases := findExistingAliases()
		// load in the lock.json
		lockDeps := loadLockDependencies()
		yamlDeps := COM.GetDependencies(false)
		deps := []string{}
		for _, dep := range yamlDeps {
			if !slices.Contains(lockDeps, dep) {
				deps = append(deps, dep)
			}
		}
		ds := strings.Split(strings.TrimPrefix(strings.TrimSuffix(strings.Join(os.Args[2:], " "), ","), ","), ",")
		for _, d := range ds {
			if !slices.Contains(yamlDeps, d) {
				COM.AddToSection("dependencies", COM.NormalizeSpaces(d))
			}
		}
		ds = append(deps, ds...)
		//ds = strings.ReplaceAll(ds, "--save-dev", "exec")
		depStrings := COM.NormalizeDependencies(ds)
		installFromYML(aliases, depStrings, true)
	}
	execChmod()
	checkUnexcluded()
	println()
	for _, v := range finishMessages {
		println("\033[38;5;208m"+tab, v, "\033[0m")
	}
}

func determineCLIClassifier(s string) any {
	if !strings.Contains(s, ":") {
		s = "*:" + s
	}
	slices := strings.SplitN(s, ":", 2)
	classifiers := COM.GetSection("classifiers", false).(map[string]string)
	if _, ok := classifiers[slices[0]]; ok {
		addFinishMessage("\033[38;5;208mWarning: Classifier for " + slices[0] + " already exists\033[0m")
		return nil
	}
	return map[string]string{slices[0]: slices[1]}
}

func execChmod() {
	entries, err := os.ReadDir("jpm_dependencies/execs")
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// TODO: turn this to Has shbang check
		if !strings.HasSuffix(name, ".jar") {
			err := os.Chmod(filepath.Join("jpm_dependencies/execs", name), 0755)
			if err != nil {
				println("Error changing permissions for", name+":", err)
			}
		}
	}
}
func QuickInstall(force bool) COM.Dependencies {
	COM.RunScript("jpm install", true)
	b, err := os.ReadFile(filepath.Join("jpm_dependencies", "lock.json"))
	if err != nil {
		os.Exit(1)
	}
	var lockDeps COM.Dependencies
	if err := json.Unmarshal(b, &lockDeps); err != nil {
		os.Exit(1)
	}
	return lockDeps
}
func installFromYML(aliases []string, deps []string, clean bool) {
	// load in locals initially
	localDeps, noneLocal := findAllLocalDeps(deps)
	fromLocal(localDeps)

	// jpm first
	jpmDeps, noneJpmdeps := findAllJPM(noneLocal, aliases)
	fromJPM(jpmDeps)

	// maven second
	repoList := getRepoList()
	repoDeps, _ := findAllRepoDeps(noneJpmdeps, repoList)
	fromRepo(repoDeps)
	installDependencies()
	if clean {
		cleanup()
	}

	//raws last
	rawDeps := findAllRaws(COM.GetDependencies(false))
	fromRAW(rawDeps)
	installScripts()
	g_lockDeps.Dependencies = COM.GetDependencies(false)
	dumpDependencies()

}

func fromLocal(localDeps []string) {
	for _, dep := range localDeps {
		println("local dependencies are not supported yet: " + dep)
		os.Exit(1)
	}
}

func findAllLocalDeps(deps []string) (localDeps []string, noneLocal []string) {
	for _, v := range deps {
		if strings.HasPrefix(v, "local ") {
			localDeps = append(localDeps, v)
		} else {
			noneLocal = append(noneLocal, v)
		}
	}
	return localDeps, noneLocal
}

func fromJPM(deps []string) string {
	depString := ""
	for _, d := range deps {
		jDep, err := disectJPMDepString(d)
		if err != nil {
			continue
		}
		currentWorkingRepo = jpmRepoUrl
		currentOuterScope = jDep.Scope
		depString = saveAllJPMSubDependencies(&jDep)
		if currentOuterScope == "exec" {
			jarfilename := jDep.Package + "-" + jDep.Version + ".jar"
			urljar := generateJpmDepUrl(currentWorkingRepo, jDep.Package, jDep.Version, jarfilename)
			urljscript := generateJpmDepUrl(currentWorkingRepo, jDep.Package, jDep.Version, jDep.Package)
			g_lockDeps.Scripts[jDep.Package] = jarfilename + "|" + urljar + "|" + urljscript
		}
	}
	addJPMSubDependenciesToDownloadList()
	return depString
}
func fromRepo(dependenciesWithRepo map[string][]Repo) {
	for _, v := range dependenciesWithRepo {
		for _, dr := range v {
			if !checkValidScope(dr.Scope) {
				printNotValidScope(dr.Scope)
				println("\033[38;5;208m" + tab + "dependency might have a faulty alias, check your package.yml\033[0m")
				continue
			} else {
				currentOuterScope = dr.Scope
			}
			currentWorkingRepo = dr.Repo
			saveAllRepoSubDependencies(&dr)
			if currentOuterScope == "exec" {
				jarfilename := dr.ArtefactID + "-" + dr.ArtVer + ".jar"
				url := generateRepoDepUrl(currentWorkingRepo, dr.GroupID, dr.ArtefactID, dr.ArtVer, jarfilename)
				g_lockDeps.Scripts[dr.ArtefactID] = jarfilename + "|" + url
			}
		}
	}

	for _, v := range getRepoList().Repos {
		currentWorkingRepo = v.Repo
		addRepoSubDependenciesToDownloadList(v.Repo)
	}
}

func fromRAW(deps []string) {
	if len(deps) == 0 {
		return
	}
	for k := range downloadInfo {
		delete(downloadInfo, k)
	}

	print("\033[32m  --- RAW: Downloading \033[0m[")
	for _, d := range deps {
		checkSlice := strings.Split(d, " ")
		if len(checkSlice) > 4 || len(checkSlice) < 2 || !strings.Contains(d, "/") {
			print("\033[31m=\033[0m")
			addFinishMessage(tab + "RAW Dependency " + d + " Does not exist")
			continue
		}
		dSlices := strings.Split(d, " ")
		scope := ""
		url := dSlices[1]
		if len(dSlices) == 3 && !strings.Contains(d, " -x ") {
			if checkValidScope(dSlices[2]) {
				scope = dSlices[2]
			} else {
				printNotValidScope(dSlices[2])
				print("\033[31m=\033[0m")
				continue
			}
		} else if len(dSlices) == 3 && strings.Contains(d, " -x ") {
			url = dSlices[2]
			scope = "extract x"
		} else if len(dSlices) == 4 {
			if checkValidScope(dSlices[3]) {
				url = dSlices[2]
				scope = dSlices[3] + " x"
			} else {
				printNotValidScope(dSlices[3])
				print("\033[31m=\033[0m")
				continue
			}
		}
		urlSlice := strings.Split(url, "/")
		filename := urlSlice[len(urlSlice)-1]
		downloadInfo[url] = []string{filename, "raw", scope}
		download(url, filename, scope, listAlreadyInstalledDeps())
		sname := strings.Split(strings.TrimSuffix(filename, filepath.Ext(filename)), "-")[0]
		if scope == "exec" {
			g_lockDeps.Scripts[sname] = filename + "|" + url
		}
	}
	println("]")
}
func findAllRaws(raws []string) []string {
	deps := []string{}
	for _, v := range raws {
		if strings.HasPrefix(COM.NormalizeSpaces(strings.ToLower(v)), "raw ") {
			found := slices.Contains(deps, v)
			if !found {
				deps = append(deps, v)
			}
		}
	}
	return deps
}
func checkValidScope(scope string) bool {
	return slices.Contains(jpmScopes, scope)
}
func printNotValidScope(scope string) {
	addFinishMessage(tab + "Warning: scope " + scope + " is a unknown scope")
}

func download(url string, filename string, scope string, depsInstalled map[string][]string) {
	sc := strings.Split(strings.Trim(scope, "|"), "|")[0]
	extract := false
	if sc == "extract x" || strings.HasSuffix(sc, " x") {
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
			if COM.Verbose {
				println(filename, "already exists, skipping download")
			}
			return
		}
		if err, _ := COM.DownloadFile(url, filepath.Join("jpm_dependencies", "tests"), filename, false, false); err != nil {
			failedInstalledList = append(failedInstalledList, tab+"Failed to correctly install : "+filename+" ERR:"+err.Error())
			print("\033[31m=\033[0m")
			return
		}
		print("=")
		if extract {
			COM.Extract(filepath.Join("jpm_dependencies", "tests"), filename)
			COM.CleanupExtract(filepath.Join("jpm_dependencies", "tests"), filename)
		}
	case "exec":
		if slices.Contains(slices.Concat(depsInstalled["exec"], depsInstalled["jpm_dependencies"]), filename) {
			if COM.Verbose {
				println(filename, "already exists, skipping download")
			}
			return
		}
		if err, _ := COM.DownloadFile(url, filepath.Join("jpm_dependencies", "execs"), filename, false, false); err != nil {
			failedInstalledList = append(failedInstalledList, tab+"Failed to correctly install : "+filename+" ERR:"+err.Error())
			print("\033[31m=\033[0m")
			return
		}
		print("=")
		if extract {
			COM.Extract(filepath.Join("jpm_dependencies", "execs"), filename)
			COM.CleanupExtract(filepath.Join("jpm_dependencies", "execs"), filename)
		}
	default:
		v, ok := depsInstalled["jpm_dependencies"]
		if ok && slices.Contains(v, filename) {
			if COM.Verbose {
				println(filename, "already exists, skipping download")
			}
			return
		}
		if err, _ := COM.DownloadFile(url, "jpm_dependencies", filename, false, false); err != nil {
			failedInstalledList = append(failedInstalledList, tab+"Failed to correctly install : "+filename+" ERR:"+err.Error())
			print("\033[31m=\033[0m")
			return
		}
		print("=")
		if extract {
			COM.Extract("jpm_dependencies", filename)
			COM.CleanupExtract("jpm_dependencies", filename)
		}
	}

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
	dir = filepath.Join("jpm_dependencies", "tests")
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
	dir = filepath.Join("jpm_dependencies", "execs")
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
	depMap := make(map[string]string, len(depsList[currentWorkingRepo])/2)

	for i := 0; i < len(depsList[currentWorkingRepo]); i += 2 {
		if depMap[depsList[currentWorkingRepo][i]] == "" {
			depMap[depsList[currentWorkingRepo][i]] = depsList[currentWorkingRepo][i+1]
		} else if depsList[currentWorkingRepo][i+1] == "" {
			continue
		} else if depMap[depsList[currentWorkingRepo][i]] < depsList[currentWorkingRepo][i+1] {
			depMap[depsList[currentWorkingRepo][i]] = depsList[currentWorkingRepo][i+1]
		}
	}
	if currentWorkingRepo != jpmRepoUrl {
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
	}
	for k := range depMap {
		if strings.HasSuffix(k, "|test") {
			d := strings.TrimSuffix(k, "test")
			if _, ok := depMap[d]; ok {
				if depMap[d] <= depMap[k] {
					depMap[d] = depMap[k]
				}
				depMap[k] = ""
			}
		}
		if strings.HasSuffix(k, "|exec") {
			d := strings.TrimSuffix(k, "exec")
			if _, ok := depMap[d]; ok {
				if depMap[d] <= depMap[k] {
					depMap[d] = depMap[k]
				}
				depMap[k] = ""
			}
		}
	}
	maps.DeleteFunc(depMap, func(k string, v string) bool { return v == "" || k == "" || strings.HasPrefix(k, "|") })
	maps.DeleteFunc(depMap, func(k string, v string) bool {
		for _, ex := range excludes {
			k = strings.TrimSuffix(k, "|test")
			k = strings.TrimSuffix(k, "|")
			ks := strings.Split(k, "|")
			k = ks[len(ks)-1]
			if k == ex {
				return true
			}
		}
		return false
	})
	switch currentWorkingRepo {
	case jpmRepoUrl:
		g_lockDeps.JPM = depMap
	default:
		maps.DeleteFunc(g_lockDeps.Repos, func(k string, v map[string]string) bool { return k == "" })
		g_lockDeps.Repos[currentWorkingRepo] = depMap
	}
	return depMap
}
func loadLockDependencies() []string {
	file, err := os.Open(filepath.Join("jpm_dependencies", "lock.json"))
	classified := false
	if COM.GetSection("classified", false) != nil {
		classified = COM.GetSection("classified", false).(bool)
	}
	if err != nil {
		g_lockDeps = COM.Dependencies{
			Classified:   classified,
			Dependencies: COM.GetDependencies(false),
			JPM:          map[string]string{},
			Repos:        map[string]map[string]string{},
			Scripts:      map[string]string{},
		}
		return []string{}
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var lockDeps COM.Dependencies
	if err := decoder.Decode(&lockDeps); err != nil {
		println("Error decoding lock.json:", err)
		os.Exit(1)
	}
	g_lockDeps = lockDeps
	if g_lockDeps.Scripts == nil {
		g_lockDeps.Scripts = map[string]string{}
	}
	if g_lockDeps.JPM == nil {
		g_lockDeps.JPM = map[string]string{}
	}
	if g_lockDeps.Repos == nil {
		g_lockDeps.Repos = map[string]map[string]string{}
	}
	if g_lockDeps.Dependencies == nil {
		g_lockDeps.Dependencies = []string{}
	}
	for k, v := range lockDeps.JPM {
		depsList[jpmRepoUrl] = append(depsList[jpmRepoUrl], k, v)
	}
	for k, v := range lockDeps.Repos {
		for k2, v2 := range v {
			depsList[k] = append(depsList[k], k2, v2)
		}
	}
	for i, d := range lockDeps.Dependencies {
		g_lockDeps.Dependencies[i] = COM.NormalizeSpaces(d)
	}
	return g_lockDeps.Dependencies
}
func installDependencies() {
	already := listAlreadyInstalledDeps()
	if len(downloadInfo) != 0 {
		print("      Downloading [")
		var wg sync.WaitGroup
		for k, v := range downloadInfo {
			wg.Add(1)
			go func(url, filename, scope string) {
				defer wg.Done()
				download(url, filename, scope, already)
			}(k, v[0], v[1])
		}
		wg.Wait()
		println("]")
	}
}
func installScripts() {
	already := listAlreadyInstalledDeps()
	if len(g_lockDeps.Scripts) != 0 {
		print("      Downloading Scripts [")
		var wg sync.WaitGroup
		for k, v := range g_lockDeps.Scripts {
			wg.Add(1)
			go func(ky, value string) {
				defer wg.Done()
				parts := strings.Split(value, "|")
				download(parts[1], parts[0], "exec", already)
				if len(parts) == 3 {
					download(parts[2], ky, "exec", already)
				}
			}(k, v)
		}
		wg.Wait()
		for k, v := range g_lockDeps.Scripts {
			parts := strings.Split(v, "|")
			createExecScript(k, parts[0])
		}
		println("]")
	}
	if len(failedInstalledList) > 0 {
		for _, v := range failedInstalledList {
			if COM.Verbose {
				println(v)
			}
		}
	}
	if !COM.Verbose {
		println("\n\033[38;5;208m Use 'jpm install -v' for more details.\033[0m")
	}
}
func dumpDependencies() {
	lockFilePath := filepath.Join("jpm_dependencies", "lock.json")

	// Remove the file if it already exists to ensure we override it
	if _, err := os.Stat(lockFilePath); err == nil {
		if err := os.Remove(lockFilePath); err != nil {
			fmt.Println("Error removing existing lock.json:", err)
			return
		}
	}

	file, err := os.Create(lockFilePath)
	if err != nil {
		fmt.Println("Error creating lock.json:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(g_lockDeps); err != nil {
		fmt.Println("Error encoding lock.json:", err)
	}
}
func cleanup() {
	jars := []string{}
	execjars := []string{}
	testjars := []string{}

	for key, v := range g_lockDeps.JPM {
		jar := ""
		if strings.HasSuffix(key, "|test") {
			jar = "tests"
		}
		if strings.HasSuffix(key, "|exec") {
			jar = "execs"
		}
		value := strings.TrimSuffix(key, "|test")
		value = strings.TrimSuffix(value, "|exec")
		valueS := strings.Split(strings.TrimSuffix(value, "|"), "|")
		value = valueS[len(valueS)-1]
		classifier := ""
		if len(valueS) == 2 {
			classifier = "-" + valueS[0]
		}
		switch jar {
		case "tests":
			testjars = append(testjars, value+"-"+v+classifier+".jar")
		case "execs":
			execjars = append(execjars, value+"-"+v+classifier+".jar", value)
		default:
			jars = append(jars, value+"-"+v+classifier+".jar")
		}
	}

	for _, depMap := range g_lockDeps.Repos {
		for key, v := range depMap {
			jar := ""
			if strings.HasSuffix(key, "|test") {
				jar = "tests"
			}
			if strings.HasSuffix(key, "|exec") {
				jar = "execs"
			}
			value := strings.TrimSuffix(key, "|test")
			value = strings.TrimSuffix(value, "|exec")
			valueS := strings.Split(strings.TrimSuffix(value, "|"), "|")
			value = valueS[len(valueS)-1]
			classifier := ""
			if len(valueS) == 3 {
				classifier = "-" + valueS[0]
			}
			switch jar {
			case "tests":
				jar = value + "-" + v + classifier + ".jar"
				testjars = append(testjars, jar)
			case "execs":
				jar = value + "-" + v + classifier + ".jar"
				execjars = append(execjars, jar, value)
			default:
				jar = value + "-" + v + classifier + ".jar"
				jars = append(jars, jar)
			}
		}
	}

	//clean up
	files, err := os.ReadDir("jpm_dependencies")
	if err == nil {
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".jar") {
				if !slices.Contains(jars, file.Name()) {
					if file.Name() == "annotations-13.0.jar" || file.Name() == "kotlin-stdlib.jar" || file.Name() == "kotlin-reflect.jar" {
						continue
					}
					os.Remove(filepath.Join("jpm_dependencies", file.Name()))
				}
			}
		}
	}
	files, err = os.ReadDir(filepath.Join("jpm_dependencies", "tests"))
	if err == nil {
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".jar") {
				if !slices.Contains(testjars, file.Name()) {
					if file.Name() == "junit.jar" || file.Name() == "kotlin-test.jar" {
						continue
					}
					os.Remove(filepath.Join("jpm_dependencies", "tests", file.Name()))
				}
			}
		}
	}

	files, err = os.ReadDir(filepath.Join("jpm_dependencies", "execs"))
	if err == nil {
		for _, file := range files {
			if !slices.Contains(execjars, file.Name()) {
				os.Remove(filepath.Join("jpm_dependencies", "execs", file.Name()))
			}
		}
	}
}
func addFinishMessage(s string) {
	finishMessages = append(finishMessages, s)
}
func GetDependenciesJson() []byte {
	loadLockDependencies()
	depsJson, err := json.MarshalIndent(g_lockDeps, "", "  ")
	if err != nil {
		return nil
	}
	return depsJson
}

var excluded map[string]bool = map[string]bool{}

func foundExcluded(s string) {
	excluded[s] = true
}
func checkUnexcluded() {
	for _, v := range excludes {
		if !excluded[v] {
			addFinishMessage("\033[33mInfo : was not excluded anywhere " + v + " in this install process\033[0m")
		}
	}
}

func createExecScript(scriptName, filename string) {
	// this does not account for classified jars.... if someone
	// is trying to execute a classified jar... maybe it should be handled here?
	// but that person can always change the name of the Main-Class in the script anyway.
	if !strings.HasSuffix(filename, ".jar") {
		return
	}
	_, err1 := os.Stat(filepath.Join("jpm_dependencies", "execs", filename))
	_, err2 := os.Stat(filepath.Join("jpm_dependencies", filename))
	if err1 != nil || err2 != nil {
		if _, err := os.Stat(filepath.Join("jpm_dependencies", "execs", scriptName)); err == nil {
			if COM.Verbose {
				println(scriptName, "already exists, skipping script creation")
			}
			return
		}
		separator := ":"
		if COM.IsWindows() {
			separator = ";"
		}
		if err1 == nil {
			filename = "jpm_dependencies/execs/" + filename
		} else {
			filename = "jpm_dependencies/" + filename
		}

		scriptCmd := `
	mainc=$(unzip -p ` + filename + ` META-INF/MANIFEST.MF | grep Main-Class | awk '/Main-Class:/ {print $2}'|tr -d '\n'| tr -d '\r')
	if [ -z "$mainc" ]; then
		jpmdir=` + COM.HomeDirUnix() + `
		classes=$(java -jar ${jpmdir}/libs/find-main.jar ` + filename + `)
		if [ $(echo $classes | wc -w) -gt 1 ]; then
			for main in $classes; do
				mainc=$main
				scriptname=` + scriptName + `-$(echo $main | awk -F. '{print $NF}')
				echo "#!/bin/sh" > jpm_dependencies/execs/$scriptname
				echo "# use -p <module_path> to add modules paths" >> jpm_dependencies/execs/$scriptname
				printf "java -cp \"jpm_dependencies/*` + separator + `jpm_dependencies/execs/*\" %s %s" "$mainc" '$@' >> jpm_dependencies/execs/$scriptname
			done
			exit 0
		elif [ $(echo $classes | wc -w) -eq 1 ]; then
			echo "#!/bin/sh" > jpm_dependencies/execs/` + scriptName + `
			echo "# use -p <module_path> to add modules paths" >> jpm_dependencies/execs/` + scriptName + `
			printf "java -cp \"jpm_dependencies/*` + separator + `jpm_dependencies/execs/*\" %s %s" "$classes" '$@' >> jpm_dependencies/execs/` + scriptName + `
			exit 0
		fi
		exit 1
	else
		echo "#!/bin/sh" > jpm_dependencies/execs/` + scriptName + `
		echo "# use -p <module_path> to add modules paths" >> jpm_dependencies/execs/` + scriptName + `
		printf "java -cp \"jpm_dependencies/*` + separator + `jpm_dependencies/execs/*\" %s %s" "$mainc" '$@' >> jpm_dependencies/execs/` + scriptName + `
	fi`
		err := COM.RunScript(scriptCmd, true)
		if err != nil {
			print("\033[33m=\033[0m")
			delete(g_lockDeps.Scripts, scriptName)
			if COM.Verbose {
				addFinishMessage("\033[33mNo Main-Class in " + filename + "\033[0m")
			}
			return
		}
		print("=")
	}
}
