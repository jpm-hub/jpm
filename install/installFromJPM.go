package install

import (
	"encoding/json"
	"errors"
	"fmt"
	COM "jpm/common"
	"slices"
	"strings"
)

type jpmRepo struct {
	Package string
	Version string
	Scope   string
}

type LastestJPM struct {
	Latest  string `json:"latest"`
	Release string `json:"release"`
}

func findAllJPM(unfiltteredDeps []string, aliases []string) (jpmDeps []string, noneJpmdeps []string) {
	jpmDeps = []string{}
	noneJpmdeps = []string{}
	for _, v := range unfiltteredDeps {
		v = COM.NormalizeSpaces(v)
		spaces := strings.Count(v, " ")
		splits := strings.SplitN(v, " ", 2)
		isInAliases := slices.Contains(append(aliases, "raw"), splits[0])
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

func disectJPMDepString(d string) (jpmRepo, error) {
	dSlices := strings.Split(d, " ")
	if len(dSlices) > 2 {
		println(tab + "JPM Dependency " + d + " Does not exist")
		return jpmRepo{}, errors.New("not nil")
	}
	scope := ""
	version := ""
	if len(dSlices) > 1 {
		if checkValidScope(dSlices[1]) {
			scope = dSlices[1]
		} else {
			printNotValidScope(dSlices[1])
			return jpmRepo{}, errors.New("wrong scope")
		}
	}
	depAndVersion := strings.Split(dSlices[0], ":")
	dep := depAndVersion[0]
	if len(depAndVersion) > 1 {
		version = depAndVersion[1]
	}
	return jpmRepo{
		Package: dep,
		Scope:   scope,
		Version: version,
	}, nil
}

func saveAllJPMSubDependencies(d *jpmRepo) string {
	version := ""
	switch d.Version {
	case "latest":
		// figure out version for this jpm dependency
		v, err := figureOutLatestJPM(d.Package)
		if err != nil {
			println("\033[31m  --- JPM: Resolving " + d.Package + " ! " + "Unable to get latest version: " + err.Error() + " \033[0m\n")
			return ""
		}
		version = v

	case "":
		// figure out version for this jpm dependency
		v, err := figureOutLatestJPM(d.Package)
		if err != nil {
			println("\033[31m  --- JPM: Resolving " + d.Package + " ! " + "Unable to get latest version: " + err.Error() + " \033[0m\n")
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
	saveJPMExecToDownloadList(d.Scope, d.Package, d.Version)
	downloadDepsJPM(d)
	println("]")
	return COM.NormalizeSpaces(fmt.Sprint(d.Package + ":" + d.Version + " " + d.Scope))
}

func saveJPMExecToDownloadList(scope, dep, version string) {
	firstLetter := strings.ToLower(dep[0:1])
	if scope == "exec" {
		filename := dep
		url := currentWorkingRepo + firstLetter + "/" + dep + "/" + version + "/" + filename
		downloadInfo[url] = []string{filename, scope + "|"}
	}
}

func downloadDepsJPM(d *jpmRepo) {
	firstLetter := strings.ToLower(d.Package[0:1])
	url := jpmRepoUrl + firstLetter + "/" + d.Package + "/" + d.Version + "/dependencies.json"
	deps, err := downloadJson(url)
	if err != nil {
		addFinishMessage("\033[31m ! Unable to download " + url + " \033[0m\n")
		return
	}
	classifier, isThere := figureOutJPMClassifier(deps, *d)
	for dep, version := range deps.JPM {
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
			if checkJPMExcludes(dep) {
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
	if slices.Contains(excludes, strings.Split(dep, "|")[1]) && currentOuterScope != "exec" {
		if COM.Verbose {
			addFinishMessage("Info : excluded " + strings.Split(dep, "|")[1])
			foundExcluded(strings.Split(dep, "|")[1])
			excluded[dep] = true
		}
		return true
	}
	return false
}

func downloadJson(url string) (COM.Dependencies, error) {
	var doc COM.Dependencies
	err, content := COM.DownloadFile(url, "", "", false, true)
	if err != nil {
		return doc, err
	}
	err = json.Unmarshal([]byte(content), &doc)
	if err != nil {
		return doc, err
	}
	return doc, nil
}
func figureOutJPMClassifier(d COM.Dependencies, info jpmRepo) (string, bool) {
	if !d.Classified {
		return "", false
	}
	classifier := COM.GetSection("classifiers", false).(map[string]string)
	v, ok := classifier[info.Package]
	vs, oks := classifier["*"]
	if ok {
		return v + "|", true
	} else if oks {
		return vs + "|", true
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
	vs, oks := classifier["*"]
	if ok {
		return v + "|", true
	} else if oks {
		return vs + "|", true
	}
	if len(dSlices) == i+2 {
		addFinishMessage("Info : default clasifier -> " + dSlices[0] + " is used for " + dSlices[i+1])
		return dSlices[0] + "|", true
	}
	addFinishMessage("Warning : a classifier needs to be specified for " + dSlices[i+1])
	return "", true
}
func figureOutJPMVersion(version string) string {
	return version
}
func addJPMSubDependenciesToDownloadList() {
	currentWorkingRepo = jpmRepoUrl
	for dep, version := range resolveDependecy() {
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
		firstLetter := strings.ToLower(dep[0:1])
		filename := dep + "-" + version + classifier + ".jar"
		url := currentWorkingRepo + firstLetter + "/" + dep + "/" + version + "/" + filename
		downloadInfo[url] = []string{filename, scope + "|"}
	}
	for k := range depsList {
		if k == jpmRepoUrl {
			continue
		}
		currentWorkingRepo = k
		addRepoSubDependenciesToDownloadList(k)
	}
}
func figureOutLatestJPM(p string) (string, error) {
	var doc LastestJPM
	firstLetter := strings.ToLower(p[0:1])
	vesionningUrl := fmt.Sprintf("%s/%s/versions.json", firstLetter, p)
	vesionningUrl = currentWorkingRepo + strings.ReplaceAll(vesionningUrl, "//", "/")
	err, content := COM.DownloadFile(vesionningUrl, "", "", false, true)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal([]byte(content), &doc)
	if err != nil {
		return "", err
	}
	if doc.Release == "" {
		return doc.Latest, nil
	}
	return doc.Release, nil
}
