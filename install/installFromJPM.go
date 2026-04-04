package install

import (
	"encoding/json"
	"errors"
	"fmt"
	COM "jpm/common"
	"slices"
	"strings"
)

type jpmDependency struct {
	Type       string
	GhUsername string
	Alias      string
	Package    string
	Version    string
	Scope      string
	RepoUrl    string
}

type LastestJPM struct {
	Latest   string `json:"latest"`
	Release  string `json:"release"`
	Redirect bool   `json:"redirect"`
}

func makeJsonFileName(pack, ver string) string {
	return fmt.Sprintf("jpm.%s-%s.json", pack, ver)
}
func findAllJPM(unfiltteredDeps []string, aliases []string) (jpmDeps []string, noneJpmdeps []string) {
	jpmDeps = []string{}
	noneJpmdeps = []string{}
	for _, v := range unfiltteredDeps {
		v = COM.NormalizeSpaces(v)
		spaces := strings.Count(v, " ")
		splits := strings.SplitN(v, " ", 2)
		isInAliases := slices.Contains(append(aliases, "raw"), strings.ToLower(splits[0]))
		switch spaces {
		case 0:
			// inevitably a jpm dependency
			jpmDeps = append(jpmDeps, v)
		case 1:
			// could be a jpm dependency if there is no alias,and the scope is valid
			if !isInAliases && checkValidScope(splits[len(splits)-1]) {
				jpmDeps = append(jpmDeps, v)
				continue
			}
			fallthrough
		case 2:
			// could be a jpm dependency if there is no alias,and there is no default repo maybe
			if !isInAliases && !slices.Contains(aliases, "default") {
				jpmDeps = append(jpmDeps, v)
				continue
			} else if !isInAliases && slices.Contains(aliases, "default") {
				// could be a repo dependency if there is a default repo
				noneJpmdeps = append(noneJpmdeps, "default "+v)
				continue
			} else {
				// we'll see what happens... might be a raw dependency
				noneJpmdeps = append(noneJpmdeps, v)
			}
		default:
			noneJpmdeps = append(noneJpmdeps, v)
		}
	}
	return
}

func disectJPMDepString(d string) (jpmDependency, error) {
	dSlices := strings.Split(d, " ")
	if len(dSlices) > 2 {
		println(tab + "JPM Dependency " + d + " Does not exist")
		return jpmDependency{}, errors.New("not nil")
	}
	scope := ""
	version := ""
	if len(dSlices) > 1 {
		if checkValidScope(dSlices[1]) {
			scope = dSlices[1]
		} else {
			println("\033[38;5;208m" + tab + "Did you add a repository url for " + dSlices[1] + " ?\033[0m")
			printNotValidScope(dSlices[1])
			return jpmDependency{}, errors.New("wrong scope")
		}
	}
	depAndVersion := strings.Split(dSlices[0], ":")
	dep := depAndVersion[0]
	if len(depAndVersion) > 1 {
		version = depAndVersion[1]
	}
	return jpmDependency{
		Type:    "jpm",
		Package: dep,
		Scope:   scope,
		Version: version,
	}, nil
}

func saveAllJPMSubDependencies(d *jpmDependency) string {
	version := ""
	var err error
	var v string
	v, err = figureOutLatestJPM(*d)

	if v == "<redirected>" {
		// figure out version for this jpm dependency
		if err != nil {
			println("\033[31m  --- JPM: Resolving " + d.Package + " ! " + "Unable to get latest version\033[0m\n")
			return ""
		}
		return handleRedirect(d)
	}
	switch d.Version {
	case "latest":
		// figure out version for this jpm dependency
		if err != nil {
			println("\033[31m  --- JPM: Resolving " + d.Package + " ! " + "Unable to get latest version\033[0m\n")
			return ""
		}
		version = v
	case "":
		// figure out version for this jpm dependency
		if err != nil {
			println("\033[31m  --- JPM: Resolving " + d.Package + " ! " + "Unable to get latest version\033[0m\n")
			return ""
		}
		version = v
		// modify the yaml at this point
		COM.ReplaceDependency(fmt.Sprintf("%s %s", d.Package, d.Scope), fmt.Sprintf("%s:%s %s", d.Package, version, d.Scope))
	default:
		version = d.Version
	}
	d.Version = version
	println("\033[32m  --- JPM: Resolving " + d.Package + ":" + version + "\033[0m")
	print("      Resolving   [")
	saveJPMExecToDownloadList(d)
	url := generateJpmDepUrl(currentWorkingRepo, d.Package, d.Version, "dependencies.json")
	downloadDepsJPM(d, url)
	println("]")
	return COM.NormalizeSpaces(fmt.Sprint(d.Package + ":" + d.Version + " " + d.Scope))
}

func handleRedirect(d *jpmDependency) string {
	depJson, err := downloadJson(generateJpmDepUrl(d.RepoUrl, d.Package, d.Version, "dependencies.json"), d)
	if err != nil {
		println("\033[31m  --- JPM: Resolving " + d.Package + " ! " + "Unable to get redirection\033[0m\n")
		return ""
	}
	dr := dependency{
		Repo:       depJson.Redirect["repo"],
		GroupID:    depJson.Redirect["groupId"],
		ArtifactID: depJson.Redirect["artifactId"],
		Version:    d.Version,
		Scope:      d.Scope,
	}
	currentOuterScope = d.Scope
	currentWorkingRepo = depJson.Redirect["repo"]
	saveAllRepoSubDependencies(&dr, ">")
	currentWorkingRepo = jpmRepoUrl
	if d.Version == "" {
		COM.ReplaceDependency(fmt.Sprintf("%s %s", d.Package, d.Scope), fmt.Sprintf("%s:%s %s", d.Package, dr.Version, d.Scope))
	}
	return COM.NormalizeSpaces(fmt.Sprint(d.Package + ":" + dr.Version + " " + d.Scope))
}

func saveJPMExecToDownloadList(d *jpmDependency) {
	if d.Scope == "exec" {
		filename := d.Package
		switch d.Type {
		case "jpm":
			url := generateJpmDepUrl(currentWorkingRepo, d.Package, d.Version, filename)
			downloadInfo[url] = []string{filename, d.Scope + "|"}
		case "github":
			url := generateGithubDepUrl(currentWorkingRepo, d.Package, d.Version, filename)
			downloadInfo[url] = []string{filename, d.Scope + "|"}
		default:
			println("type :", d.Type, "not yet supported")
		}
	}
}
func generateJpmDepUrl(repo, pack, ver, filename string) string {
	firstLetter := strings.ToLower(pack[0:1])
	return repo + firstLetter + "/" + pack + "/" + ver + "/" + filename
}
func downloadDepsJPM(d *jpmDependency, url string) {
	if checkJPMExcludes(d.Package) {
		return
	}

	deps, err := downloadJson(url, d)
	if err != nil {
		addFinishMessage("\033[31m ! Unable to download " + url + " \033[0m\n")
		return
	}
	classifier, isThere := figureOutJPMClassifier(deps, *d)

	for dep, version := range deps.JPM[currentWorkingRepo] {
		if checkJPMExcludes(dep) {
			continue
		}
		_classifier, innerIsThere := figureOutJPMInnerClassifier(dep, false)
		_version := figureOutJPMVersion(version)
		if innerIsThere {
			dep = strings.TrimLeft(dep, _classifier)
			depsList[currentWorkingRepo] = append(depsList[currentWorkingRepo], _classifier+dep+"|"+currentOuterScope, _version)
		}
		depsList[currentWorkingRepo] = append(depsList[currentWorkingRepo], dep+"|"+currentOuterScope, _version)
		print("-")
	}

	for repo, deps := range deps.Repos {
		for dep, version := range deps {
			depS := strings.Split(dep, "|")
			if checkRepoExcludes(dependency{ArtifactID: depS[len(depS)-1], GroupID: depS[len(depS)-1]}) {
				continue
			}
			_classifier, innerIsThere := figureOutJPMInnerClassifier(dep, true)
			_version := figureOutJPMVersion(version)
			if innerIsThere {
				dep = strings.TrimLeft(dep, _classifier)
				depsList[repo] = append(depsList[repo], _classifier+dep+"|"+currentOuterScope, _version)
			}
			depsList[repo] = append(depsList[repo], dep+"|"+currentOuterScope, _version)
			print("-")
		}
	}
	depsList[currentWorkingRepo] = append(depsList[currentWorkingRepo], d.Package+"|"+currentOuterScope, d.Version)
	if isThere {
		depsList[currentWorkingRepo] = append(depsList[currentWorkingRepo], classifier+d.Package+"|"+currentOuterScope, d.Version)
	}
	print("-")
}
func checkJPMExcludes(dep string) bool {
	depS := strings.Split(dep, "|")
	dep = depS[len(depS)-1]
	if slices.Contains(excludes, dep) {
		if COM.Verbose {
			addFinishMessage("Info : excluded " + dep)
		}
		foundExcluded(dep)
		return true
	}
	return false
}

func figureOutJPMClassifier(d COM.Dependencies, info jpmDependency) (string, bool) {
	classifier := COM.GetSection("classifiers", false).(map[string]string)
	v, ok := classifier[info.Package]
	vs, oks := classifier["*"]
	if ok && d.Classified {
		return v + "|", true
	} else if oks && d.Classified {
		return vs + "|", true
	}
	if !d.Classified {
		return "", false
	}
	addFinishMessage("Warning : a classifier needs to be specified for " + info.Package)
	return "", true
}
func figureOutJPMInnerClassifier(d string, forRepos bool) (string, bool) {
	dSlices := strings.Split(d, "|")
	i := 0
	if forRepos {
		i = 1
	}
	if len(dSlices) < i+2 {
		return "", false
	}
	classifier := COM.GetSection("classifiers", false).(map[string]string)
	v, ok := classifier[dSlices[i+1]]
	vg, okg := classifier[dSlices[i]]
	vs, oks := classifier["*"]
	if ok {
		return v + "|", true
	} else if forRepos && okg {
		return vg + "|", true
	} else if oks {
		return vs + "|", true
	}
	if len(dSlices) == i+2 && dSlices[0] != "" {
		addFinishMessage("Info : default clasifier -> " + dSlices[0] + " is used for " + dSlices[i+1])
		return dSlices[0] + "|", true
	}
	addFinishMessage("Warning : a classifier needs to be specified for " + dSlices[i+1])
	return "", true
}
func figureOutJPMVersion(version string) string {
	return version
}
func addJPMSubDependenciesToDownloadList(repoUrl string, depMap map[string]string) {
	currentWorkingRepo = repoUrl
	for dep, version := range depMap {
		depS := strings.Split(dep, "|")
		classifier := ""
		scope := ""
		switch len(depS) {
		case 1:
			dep = depS[0]
		case 2:
			dep = depS[0]
			scope = depS[1]
		case 3:
			dep = depS[1]
			scope = depS[2]
			classifier = "-" + depS[0]
		}
		filename := dep + "-" + version + classifier + ".jar"
		url := generateJpmDepUrl(currentWorkingRepo, dep, version, filename)
		downloadInfo[url] = []string{filename, scope + "|"}
	}
	g_lockDeps.JPM[repoUrl] = depMap
}
func figureOutLatestJPM(p jpmDependency) (string, error) {

	var doc LastestJPM
	vesionningUrl := generateJpmDepUrl(currentWorkingRepo, p.Package, p.Version, "versions.json")
	err, content := COM.DownloadFile(vesionningUrl, "", "", false, true)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal([]byte(content), &doc)
	if err != nil {
		return "", err
	}
	if doc.Redirect == true {
		return "<redirected>", nil
	}
	if doc.Release == "" {
		return doc.Latest, nil
	}
	return doc.Release, nil
}
