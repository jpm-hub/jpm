package install

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	COM "jpm/common"
	"net/http"
	"slices"
	"strings"
)

type dependencyManagement struct {
	Dependencies []dependency
}
type dependency struct {
	GroupID    string
	ArtifactID string
	Version    string
	Scope      string
	optional   string
	Classifier string
	Type       string // e.g., "jar", "war", "aar"
}
type pom struct {
	Parent               *dependency
	DependencyManagement *dependencyManagement
	Dependencies         []dependency
	Version              string
	gid                  string
	aid                  string
	Properties           map[string]string
	Packaging            string
	Url                  string
}
type document struct {
	XMLName   xml.Name `xml:"project"`
	Version   string   `xml:"version"`
	Gid       string   `xml:"groupId"`
	Aid       string   `xml:"artifactId"`
	Packaging string   `xml:"packaging"`
	Parent    struct {
		GroupID    string `xml:"groupId"`
		ArtifactID string `xml:"artifactId"`
		Version    string `xml:"version"`
	} `xml:"parent"`
	DependencyManagement struct {
		Dependencies []struct {
			GroupID    string `xml:"groupId"`
			ArtifactID string `xml:"artifactId"`
			Version    string `xml:"version"`
			Scope      string `xml:"scope"`
			Optional   string `xml:"optional"`
			Classifier string `xml:"classifier"`
		} `xml:"dependencies>dependency"`
	} `xml:"dependencyManagement"`
	Dependencies []struct {
		GroupID    string `xml:"groupId"`
		ArtifactID string `xml:"artifactId"`
		Version    string `xml:"version"`
		Scope      string `xml:"scope"`
		Optional   string `xml:"optional"`
		Classifier string `xml:"classifier"`
	} `xml:"dependencies>dependency"`

	Properties struct {
		Value string `xml:",innerxml"`
	} `xml:"properties"`
}
type LastestRepo struct {
	Metadata   xml.Name `xml:"metadata"`
	Versioning struct {
		Latest  string `xml:"latest"`
		Release string `xml:"release"`
	} `xml:"versioning"`
}
type props struct {
	p map[string]string
}

var resolvedAlready map[string]uint8 = map[string]uint8{}

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
		url := ""
		alias := ""
		username := ""
		password := ""
		for k, v := range repoSectionMap {
			switch k {
			case "username":
				username = v
			case "password":
				password = v
			case "type":
			default:
				alias = k
				if !strings.HasSuffix(v, "/") {
					url = v + "/"
				} else {
					url = v
				}
			}
		}
		repos.Repos = append(repos.Repos, Repo{Alias: strings.ToLower(alias), Repo: url, Username: username, Password: password})
	}
	return repos
}
func findExistingAliases() []string {
	repos := getRepoList()
	aliases := []string{}
	for _, v := range repos.Repos {
		aliases = append(aliases, strings.ToLower(v.Alias))
	}
	return aliases
}
func findAllRepoDeps(deps []string, repoList Repositories) (repos map[string][]Repo, raws []string) {
	repos = map[string][]Repo{}
	for _, v := range deps {
		found := false
		for _, repoFromYaml := range repoList.Repos {
			v = COM.NormalizeSpaces(v)
			if strings.HasPrefix(v, repoFromYaml.Alias) {
				r, err := disectRepoDepString(v, repoFromYaml.Repo, repoFromYaml.Alias)
				if err == nil {
					found = true
					repos[repoFromYaml.Repo] = append(repos[repoFromYaml.Repo], r)
				}
			}
		}
		if !found {
			raws = append(raws, v)
		}
	}
	return repos, raws
}
func disectRepoDepString(depString string, repoURL string, alias string) (Repo, error) {
	dSlice := strings.Split(depString, " ")
	if len(dSlice) < 3 {
		println("\033[31m" + tab + "The dependency : " + depString + " is ambigious and will be ignored\033[0m")
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
		Repo:       repoURL,
		Alias:      alias,
		GroupID:    dSlice[1],
		ArtefactID: artver[0],
		ArtVer:     COM.NormalizeSpaces(version),
		Scope:      t,
	}, nil
}
func saveAllRepoSubDependencies(dr *Repo) error {
	version := dr.ArtVer
	var err error
	if dr.ArtVer == "" {
		version, err = figureOutLastestRepo(dr.GroupID, dr.ArtefactID)
		if err != nil {
			println("\033[31m  --- " + strings.ToUpper(dr.Alias) + ": Resolving " + dr.ArtefactID + " ! " + "Unable to get latest version: " + err.Error() + " \033[0m\n")
			return errors.New("not nil")
		}
		dr.ArtVer = version
		// modify the yaml at this point
		if dr.Alias != "default" {
			COM.ReplaceDependency(fmt.Sprintf("%s %s %s %s", dr.Alias, dr.GroupID, dr.ArtefactID, dr.Scope), fmt.Sprintf("%s %s %s:%s %s", dr.Alias, dr.GroupID, dr.ArtefactID, dr.ArtVer, dr.Scope))
		} else {
			COM.ReplaceDependency(fmt.Sprintf("%s %s %s", dr.GroupID, dr.ArtefactID, dr.Scope), fmt.Sprintf("%s %s:%s %s", dr.GroupID, dr.ArtefactID, dr.ArtVer, dr.Scope))
		}
	}
	if dr.ArtVer == "latest" {
		version, err = figureOutLastestRepo(dr.GroupID, dr.ArtefactID)
		if err != nil {
			println("\033[31m  --- " + strings.ToUpper(dr.Alias) + ": Resolving " + dr.ArtefactID + " ! " + "Unable to get latest version: " + err.Error() + " \033[0m\n")
			return errors.New("not nil")
		}
		dr.ArtVer = version
	}
	al := ">"
	if dr.Alias != "default" {
		al = strings.ToUpper(dr.Alias) + ":"
	}
	println("\033[32m  --- " + al + " Resolving " + dr.ArtefactID + ":" + version + "\033[0m")
	print("      Resolving   [")
	p := downloadDepsRepo(currentWorkingRepo, dr.GroupID, dr.ArtefactID, version, false)
	if p == nil {
		print("\033[31m-\033[0m")
	}
	figureOutAllLatestAndDownload(currentWorkingRepo)
	println("]")
	return nil
}
func addRepoSubDependenciesToDownloadList(url string) {
	for k, v := range resolveDependecy() {
		gas := strings.Split(k, "|")
		groupID := gas[0]
		artefactID := gas[1]
		classifier := ""
		if len(gas) == 4 {
			classifier = "-" + gas[0]
			groupID = gas[1]
			artefactID = gas[2]
		}
		version := v
		scope := gas[len(gas)-1]
		url := url + strings.ReplaceAll(groupID, ".", "/") + "/" + artefactID + "/" + version + "/" + artefactID + "-" + version + classifier + ".jar"
		filename := artefactID + "-" + version + classifier + ".jar"
		downloadInfo[url] = []string{filename, "" + scope + "|"}
	}
}

func parsePOM(pomContent string) pom {
	var doc document

	decoder := xml.NewDecoder(strings.NewReader(pomContent))
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		return input, nil
	}
	decoder.Decode(&doc)
	result := pom{
		Properties: make(map[string]string),
	}

	// Parse properties
	var p props
	xml.Unmarshal([]byte("<jpm.root>"+strings.TrimSpace(doc.Properties.Value)+"</jpm.root>"), &p)

	result.Properties = p.p

	// Parse parent
	if doc.Parent.GroupID != "" {
		result.Parent = &dependency{
			GroupID:    doc.Parent.GroupID,
			ArtifactID: doc.Parent.ArtifactID,
			Version:    doc.Parent.Version,
		}
	}
	if doc.Version != "" {
		result.Version = doc.Version
	}

	// Parse dependencyManagement
	if len(doc.DependencyManagement.Dependencies) > 0 {
		result.DependencyManagement = &dependencyManagement{}
		for _, dep := range doc.DependencyManagement.Dependencies {
			result.DependencyManagement.Dependencies = append(
				result.DependencyManagement.Dependencies,
				dependency{
					GroupID:    dep.GroupID,
					ArtifactID: dep.ArtifactID,
					Version:    dep.Version,
					optional:   dep.Optional,
					Scope:      dep.Scope,
					Classifier: dep.Classifier,
				},
			)
		}
	}

	// Parse dependencies
	for _, dep := range doc.Dependencies {
		result.Dependencies = append(
			result.Dependencies,
			dependency{
				GroupID:    dep.GroupID,
				ArtifactID: dep.ArtifactID,
				Version:    dep.Version,
				optional:   dep.Optional,
				Scope:      dep.Scope,
				Classifier: dep.Classifier,
			},
		)
	}

	// parse gid
	result.gid = doc.Gid
	result.Packaging = doc.Packaging
	// Parse aid
	result.aid = doc.Aid

	return result
}
func generateRepoDepUrl(repo, gid, aid, ver, filename string) string {
	gidPath := strings.ReplaceAll(gid, ".", "/")
	return repo + gidPath + "/" + aid + "/" + ver + "/" + filename
}
func downloadDepsRepo(repo string, groupID string, artifactID string, version string, scopeImport bool) *pom {
	if version == "" {
		return nil
	}
	// Download and parse POM file
	pomURL := generateRepoDepUrl(repo, groupID, artifactID, version, artifactID+"-"+version+".pom")
	dep := groupID + "|" + artifactID + "|" + currentOuterScope
	pomContent := ""
	_, ok := cache[pomURL]
	if !ok {
		var err error
		pomContent, err = downloadPOM(pomURL)
		if err != nil {
			if COM.Verbose {
				println("\n  --- " + artifactID + " " + err.Error())
			}
			return nil
		}
		// Parse POM XML
		p := parsePOM(pomContent)
		p.Url = repo
		cache[pomURL] = p
	}
	pom := cache[pomURL]
	already := resolvedAlready[groupID+"|"+artifactID] > 3
	if already {
		return &pom
	}
	if scopeImport {
		savingImports(repo, &pom)
		return &pom
	}
	if strings.ToLower(pom.Packaging) != "pom" {
		if checkRepoExcludes(artifactID) {
			return &pom
		}
		classifier, classified := figureOutRepoClassifier(dependency{
			GroupID:    groupID,
			ArtifactID: artifactID,
			Version:    version,
			Scope:      currentOuterScope,
			Classifier: "",
			Type:       "",
		})
		depsList[repo] = append(depsList[repo], dep, version)
		if classified {
			depsList[repo] = append(depsList[repo], classifier+dep, version)
		}
	}
	for _, dep := range pom.Dependencies {
		scope := figureOutScope(repo, dep, pom)
		dep.Scope = scope
		optional := figureOutOptional(repo, dep, pom)
		// if scope == import add to import depsManagement list

		if (optional == "" || strings.ToLower(optional) == "false") && !strings.HasPrefix(groupID, "org.junit") {
			groupid := figureOutGroupID(repo, dep, pom)
			artifactid := figureOutArtifactID(repo, dep, pom)
			if slices.Contains(scopesAccepted, scope) {
				dep.GroupID = groupid
				dep.ArtifactID = artifactid
				version := figureOutVersion(repo, dep, pom)
				dep.Version = version
				classifier, classified := figureOutRepoClassifier(dep)
				dep.Classifier = classifier
				if checkRepoExcludes(artifactid) {
					continue
				}
				resolvedAlready[groupID+"|"+artifactID] += 1
				// parallel download maybe??
				p := downloadDepsRepo(repo, groupid, artifactid, version, false)
				if p != nil {
					depsList[p.Url] = append(depsList[p.Url], groupid+"|"+artifactid+"|"+currentOuterScope, version)
					if classified {
						depsList[p.Url] = append(depsList[p.Url], classifier+groupid+"|"+artifactid+"|"+currentOuterScope, version)
					}
					continue
				}
				if otherP := checkOtherRepositories(dep); otherP != nil {
					depsList[otherP.Url] = append(depsList[otherP.Url], groupid+"|"+artifactid+"|"+currentOuterScope, version)
					if classified {
						depsList[otherP.Url] = append(depsList[otherP.Url], classifier+groupid+"|"+artifactid+"|"+currentOuterScope, version)
					}
					continue
				}
				latests = append(latests, groupid+"|"+artifactid)
			}
		}

	}
	return &pom
}

func checkOtherRepositories(dep dependency) *pom {
	for _, repo := range getRepoList().Repos {
		if repo.Repo == currentWorkingRepo {
			continue
		}
		rp := repo.Repo
		p := downloadDepsRepo(rp, dep.GroupID, dep.ArtifactID, dep.Version, false)
		if p != nil {
			p.Url = rp
			return p
		}
	}
	return nil
}

func checkRepoExcludes(artifactID string) bool {
	for _, ex := range excludes {
		if strings.Contains(artifactID, ex) {
			if COM.Verbose {
				addFinishMessage("Info : excluded " + artifactID)
				foundExcluded(artifactID)
			}
			excluded[artifactID] = true
			return true
		}
	}
	return false
}
func savingImports(repo string, p *pom) {
	for _, dm := range p.DependencyManagement.Dependencies {
		groupid := figureOutGroupID(repo, dm, *p)
		artifactid := figureOutArtifactID(repo, dm, *p)
		version := figureOutVersion(repo, dm, *p)
		dep := groupid + "|" + artifactid + "|" + currentOuterScope
		importList = append(importList, dep, version)
	}

}

func checkParentProperty(repo string, p pom, propertyName string) string {

	if p.Parent != nil {
		parentPom := downloadDepsRepo(repo, p.Parent.GroupID, p.Parent.ArtifactID, p.Parent.Version, false)
		if parentPom != nil {
		recheckProp:
			if value, ok := p.Properties[propertyName]; ok {
				if strings.HasPrefix(value, "${") {
					propertyName = strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")
					goto recheckProp
				}
				return value
			}
			return checkParentProperty(repo, *parentPom, propertyName)
		}
	}
	return ""
}

func checkParentDependencyManagement(repo string, pom pom, groupID string, artifactID string) (string, string, string, *pom) {
	if pom.Parent == nil {
		return "", "", "", nil
	}
	parent := downloadDepsRepo(repo, pom.Parent.GroupID, pom.Parent.ArtifactID, pom.Parent.Version, false)
	if parent != nil && parent.DependencyManagement != nil {
		imports := []string{}
	afterImports:
		for _, managedDep := range parent.DependencyManagement.Dependencies {
			if managedDep.Scope == "import" && !slices.Contains(imports, managedDep.GroupID+"|"+managedDep.ArtifactID+"|"+currentOuterScope) {
				imports = append(imports, managedDep.GroupID+"|"+managedDep.ArtifactID+"|"+currentOuterScope)
				version := figureOutVersion(repo, managedDep, *parent)
				p := downloadDepsRepo(repo, managedDep.GroupID, managedDep.ArtifactID, version, true)
				if p != nil && p.DependencyManagement != nil {
					parent.DependencyManagement.Dependencies = append(parent.DependencyManagement.Dependencies, p.DependencyManagement.Dependencies...)
				}
				goto afterImports
			}
			if managedDep.GroupID == groupID && managedDep.ArtifactID == artifactID {
				return managedDep.Scope, managedDep.optional, managedDep.Version, parent
			}
		}
	}
	if parent == nil {
		return "", "", "", nil
	}
	return checkParentDependencyManagement(repo, *parent, groupID, artifactID)
}

func figureOutVersion(repo string, depwithversion dependency, p pom) string {
	// first check if optional is empty
	if depwithversion.Version == "" {
		// if empty check in dependencyManagement
		if p.DependencyManagement != nil {
			for _, managedDep := range p.DependencyManagement.Dependencies {
				if managedDep.GroupID == depwithversion.GroupID && managedDep.ArtifactID == depwithversion.ArtifactID {
					return figureOutVersion(repo, managedDep, p)
				}
			}
		}
		// if not found check parent's dependencyManagement
		if _, _, version, parent := checkParentDependencyManagement(repo, p, depwithversion.GroupID, depwithversion.ArtifactID); version != "" {
			depwithversion.Version = version
			return figureOutVersion(repo, depwithversion, *parent)
		}
		return ""
	}
	if strings.HasPrefix(depwithversion.Version, "${") {
		if strings.Contains(depwithversion.Version, "${project.version}") {
			if p.Version == "" {
				return p.Parent.Version
			} else if strings.HasPrefix(p.Version, "${") {
				depwithversion.Version = p.Version
			} else {
				return p.Version
			}
		}
		propertyName := strings.TrimSuffix(strings.TrimPrefix(depwithversion.Version, "${"), "}")

		// check for version in pom.Propreties
	recheckPropSon:
		if strings.Contains(propertyName, "project.version") {
			if p.Version == "" {
				return p.Parent.Version
			} else {
				return p.Version
			}
		}
		if value, ok := p.Properties[propertyName]; ok {
			if strings.HasPrefix(value, "${") {
				propertyName = strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")
				goto recheckPropSon
			}
			return value
		}

		// if not found check for parent by using downloadDepsRepo(parentGroupID,parentArtifactID,parentVersion)
		value := checkParentProperty(repo, p, propertyName)
		if value != "" {
			return value
		}
		return ""
	}

	return depwithversion.Version
}

func figureOutScope(repo string, depwithscope dependency, pom pom) string {
	// first check if optional is empty
	if depwithscope.Scope == "" {
		// if empty check in dependencyManagement
		if pom.DependencyManagement != nil {
			for _, managedDep := range pom.DependencyManagement.Dependencies {
				if managedDep.GroupID == depwithscope.GroupID && managedDep.ArtifactID == depwithscope.ArtifactID {
					if managedDep.Scope == "" {
						break
					}
					return managedDep.Scope
				}
			}
		}

		// if not found check parent's dependencyManagement
		if scope, _, _, p := checkParentDependencyManagement(repo, pom, depwithscope.GroupID, depwithscope.ArtifactID); scope != "" {
			depwithscope.Scope = scope
			if scope == "" {
				return ""
			}
			return figureOutScope(repo, depwithscope, *p)
		}
	}

	if strings.HasPrefix(depwithscope.Scope, "${") {
		propertyName := strings.TrimSuffix(strings.TrimPrefix(depwithscope.Scope, "${"), "}")

		// check for version in pom.Propreties
		if value, ok := pom.Properties[propertyName]; ok {
			if strings.HasPrefix(value, "${") {
				// println(" is this circular ?")
				return checkParentProperty(repo, pom, propertyName)
			}
			return value
		}

		value := checkParentProperty(repo, pom, propertyName)
		if value != "" {
			return value
		}
		// if not found return ""
		return ""
	}

	// return optional
	return depwithscope.Scope
}

func figureOutOptional(repo string, depwithoptional dependency, pom pom) string {
	// first check if optional is empty
	if depwithoptional.optional == "" {
		// if empty check in dependencyManagement
		if pom.DependencyManagement != nil {
			for _, managedDep := range pom.DependencyManagement.Dependencies {
				if managedDep.GroupID == depwithoptional.GroupID && managedDep.ArtifactID == depwithoptional.ArtifactID {
					if managedDep.optional == "" {
						break
					}
					return figureOutOptional(repo, managedDep, pom)
				}
			}
		}

		// if not found check parent's dependencyManagement
		if _, optional, _, p := checkParentDependencyManagement(repo, pom, depwithoptional.GroupID, depwithoptional.ArtifactID); optional != "" {
			depwithoptional.Scope = optional
			if optional == "" {
				return ""
			}
			return figureOutScope(repo, depwithoptional, *p)
		}
	}

	if strings.HasPrefix(depwithoptional.optional, "${") {
		propertyName := strings.TrimSuffix(strings.TrimPrefix(depwithoptional.optional, "${"), "}")

		// check for version in pom.Propreties
		if value, ok := pom.Properties[propertyName]; ok {
			if strings.HasPrefix(value, "${") {
				// println(" is this circular ?")
				return checkParentProperty(repo, pom, propertyName)
			}
			return value
		}

		value := checkParentProperty(repo, pom, propertyName)
		if value != "" {
			return value
		}
		// if not found return ""
		return ""
	}

	// return optional
	return depwithoptional.optional
}

func figureOutGroupID(repo string, depwithgroupid dependency, pom pom) string {
	if strings.Contains(depwithgroupid.GroupID, "${project.groupId}") {
		if pom.gid == "" {
			return pom.Parent.GroupID
		} else if strings.HasPrefix(pom.gid, "${") {
			depwithgroupid.GroupID = pom.gid
		} else {
			return pom.gid
		}
	}
	// first check if groupid is ${}
	if strings.HasPrefix(depwithgroupid.GroupID, "${") {
		propertyName := strings.TrimSuffix(strings.TrimPrefix(depwithgroupid.GroupID, "${"), "}")
		if depwithgroupid.GroupID == "${project.groupId}" {
			if pom.gid == "" {
				return pom.Parent.GroupID
			} else {
				return pom.gid
			}
		}
		// check for groupid in pom.Propreties
		if value, ok := pom.Properties[propertyName]; ok {
			return value
		}

		if value := checkParentProperty(repo, pom, propertyName); value != "" {
			return value
		}

		// if not found return ""
		return ""
	}

	// if found return groupid
	return depwithgroupid.GroupID
}

func figureOutArtifactID(repo string, depwithartifactid dependency, pom pom) string {
	// first check if artifactid is ${}
	if strings.HasPrefix(depwithartifactid.ArtifactID, "${") {
		propertyName := strings.TrimSuffix(strings.TrimPrefix(depwithartifactid.ArtifactID, "${"), "}")
		if depwithartifactid.ArtifactID == "${project.artifactId}" {
			return pom.gid
		}
		// check for artifactid in pom.Propreties
		if value, ok := pom.Properties[propertyName]; ok {
			return value
		}

		if value := checkParentProperty(repo, pom, propertyName); value != "" {
			return value
		}

		// if not found return ""
		return ""
	}

	// if found return artifactid
	return depwithartifactid.ArtifactID
}

func figureOutRepoClassifier(dep dependency) (string, bool) {
	classDep := COM.GetSection("classifiers", false).(map[string]string)
	vals, oks := classDep["*"]
	valA, okA := classDep[dep.ArtifactID]
	valG, okG := classDep[dep.GroupID]
	if okA {
		return valA + "|", true
	} else if okG {
		return valG + "|", true
	} else if oks {
		return vals + "|", true
	}
	if dep.Classifier == "" {
		return "", false
	}
	if !strings.HasPrefix(dep.Classifier, "${") {
		addFinishMessage("Info : default classifier -> " + dep.Classifier + " is used for " + dep.ArtifactID)
		return dep.Classifier + "|", true
	}
	c := strings.TrimSuffix(strings.TrimPrefix(dep.Classifier, "${"), "}")
	if _, ok := resolvedAlready[dep.GroupID+"|"+dep.ArtifactID]; !ok {
		addFinishMessage("Warning : classifier -> " + c + " needs to be specified for " + dep.ArtifactID)
	}
	return "", true
}

func figureOutAllLatestAndDownload(repo string) {
	for i, s := range latests {
		if i == 0 {
			latests = []string{}
		}
		gas := strings.Split(s, "|")
		groupID := gas[0]
		artifactID := gas[1]
		err := figureOutLatestAndDownload(repo, groupID, artifactID)
		if err != nil {
			addFinishMessage(err.Error())
		}
	}
	if len(latests) != 0 {
		figureOutAllLatestAndDownload(repo)
	}

}
func figureOutLatestAndDownload(repo string, groupID string, artifactID string) error {
	version, err := figureOutLastestRepo(groupID, artifactID)
	if err != nil {
		print("\033[31m-\033[0m")
		return err
	}
	p := downloadDepsRepo(repo, groupID, artifactID, version, false)
	if p == nil {
		print("\033[31m-\033[0m")
		return errors.New(tab + " failed to resolve :" + groupID + " " + artifactID + " " + version)
	}
	return nil
}
func figureOutLastestRepo(groupID string, artifactID string) (string, error) {
	var doc LastestRepo
	vesionningUrl := fmt.Sprintf("%s/%s/maven-metadata.xml", strings.ReplaceAll(groupID, ".", "/"), artifactID)
	vesionningUrl = currentWorkingRepo + strings.ReplaceAll(vesionningUrl, "//", "/")
	err, content := COM.DownloadFile(vesionningUrl, "", "", false, true)
	if err != nil {
		return "", errors.New("could not get latest for " + artifactID + " " + err.Error())
	}
	err = xml.Unmarshal([]byte(content), &doc)
	if err != nil {
		return "", errors.New("could not get latest for " + artifactID + " " + err.Error())
	}
	if doc.Versioning.Release == "" {
		return doc.Versioning.Latest, nil
	}
	return doc.Versioning.Release, nil
}
func downloadPOM(pomURL string) (string, error) {
	resp, err := http.Get(pomURL)
	if err != nil {
		return "", fmt.Errorf("failed to download POM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("POM: status code %d trying other repos if any", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read POM content: %w", err)
	}
	print("-")
	return string(body), nil
}

func (x *props) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	x.p = map[string]string{}
	last := ""
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		switch elem := token.(type) {
		case xml.EndElement:
			x.p[elem.Name.Local] = last
		case xml.CharData:
			last = strings.TrimSpace(string(elem))
		}
	}
}
