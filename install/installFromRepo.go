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
	scope      string
	optional   string
}
type pom struct {
	Parent               *dependency
	DependencyManagement *dependencyManagement
	Dependencies         []dependency
	Version              string
	gid                  string
	aid                  string
	Properties           map[string]string
	packaging            string
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
		} `xml:"dependencies>dependency"`
	} `xml:"dependencyManagement"`
	Dependencies []struct {
		GroupID    string `xml:"groupId"`
		ArtifactID string `xml:"artifactId"`
		Version    string `xml:"version"`
		Scope      string `xml:"scope"`
		Optional   string `xml:"optional"`
	} `xml:"dependencies>dependency"`

	Properties struct {
		Value string `xml:",innerxml"`
	} `xml:"properties"`
}

func parsePOM(pomContent string) pom {
	var doc document

	xml.Unmarshal([]byte(pomContent), &doc)
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
					scope:      dep.Scope,
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
				scope:      dep.Scope,
			},
		)
	}

	// parse gid
	result.gid = doc.Gid
	result.packaging = doc.Packaging
	// Parse aid
	result.aid = doc.Aid

	return result
}

func downloadDepsRepo(repo string, groupID string, artifactID string, version string, scopeImport bool) *pom {
	currentWorkingRepo = repo
	// Download and parse POM file
	pomURL := fmt.Sprintf("%s/%s/%s/%s-%s", strings.ReplaceAll(groupID, ".", "/"), artifactID, version, artifactID, version)
	pomURL = repo + strings.ReplaceAll(pomURL, "//", "/")
	dep := groupID + "|" + artifactID + "|" + currentOuterScope
	pomContent := ""
	if _, ok := cache[pomURL]; !ok {
		var err error
		pomContent, err = downloadPOM(pomURL + ".pom")
		if err != nil {
			return nil
		}
		// Parse POM XML
		cache[pomURL] = parsePOM(pomContent)
	}
	pom := cache[pomURL]
	if scopeImport {
		savingImports(&pom)
		return &pom
	}
	if strings.ToLower(pom.packaging) != "pom" {
		depsList = append(depsList, dep, version)
	}

	for _, dep := range pom.Dependencies {
		scope := figureOutScope(dep, pom)
		optional := figureOutOptional(dep, pom)
		// if scope == import add to import depsManagement list

		if (optional == "" || strings.ToLower(optional) == "false") && slices.Contains(scopesAccepted, scope) && !strings.HasPrefix(groupID, "org.junit") {
			groupid := figureOutGroupID(dep, pom)
			if strings.HasPrefix(groupid, "org.junit") {
				continue
			}
			dep.GroupID = groupid
			artifactid := figureOutArtifactID(dep, pom)
			dep.ArtifactID = artifactid
			version := figureOutVersion(dep, pom)
			depsList = append(depsList, groupid+"|"+artifactid+"|"+currentOuterScope, version)
			p := downloadDepsRepo(repo, groupid, artifactid, version, false)
			if p != nil {
				continue
			}
			latests = append(latests, groupid+"|"+artifactid)
			println(tab + " failed to have :" + groupid + " " + artifactid + " " + version + " from: " + pomURL)
		}
	}
	return &pom
}

func savingImports(p *pom) {
	for _, dm := range p.DependencyManagement.Dependencies {
		groupid := figureOutGroupID(dm, *p)
		artifactid := figureOutArtifactID(dm, *p)
		version := figureOutVersion(dm, *p)
		dep := groupid + "|" + artifactid + "|" + currentOuterScope
		importList = append(importList, dep, version)
	}

}

func checkParentProperty(p pom, propertyName string) string {

	if p.Parent != nil {
		parentPom := downloadDepsRepo(currentWorkingRepo, p.Parent.GroupID, p.Parent.ArtifactID, p.Parent.Version, false)
		if parentPom != nil {
		recheckProp:
			if value, ok := p.Properties[propertyName]; ok {
				if strings.HasPrefix(value, "${") {
					propertyName = strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")
					goto recheckProp
				}
				return value
			}
			return checkParentProperty(*parentPom, propertyName)
		}
	}
	return ""
}

func checkParentDependencyManagement(pom pom, groupID string, artifactID string) (string, string, string, *pom) {
	// time.Sleep(time.Millisecond * 800)
	// println("got into :" + pom.gid + " " + pom.aid)
	if pom.Parent == nil {
		return "", "", "", nil
	}
	parent := downloadDepsRepo(currentWorkingRepo, pom.Parent.GroupID, pom.Parent.ArtifactID, pom.Parent.Version, false)
	if parent != nil && parent.DependencyManagement != nil {
		imports := []string{}
	afterImports:
		for _, managedDep := range parent.DependencyManagement.Dependencies {
			if managedDep.scope == "import" && !slices.Contains(imports, managedDep.GroupID+"|"+managedDep.ArtifactID+"|"+currentOuterScope) {
				imports = append(imports, managedDep.GroupID+"|"+managedDep.ArtifactID+"|"+currentOuterScope)
				version := figureOutVersion(managedDep, *parent)
				p := downloadDepsRepo(currentWorkingRepo, managedDep.GroupID, managedDep.ArtifactID, version, true)
				if p != nil && p.DependencyManagement != nil {
					parent.DependencyManagement.Dependencies = append(parent.DependencyManagement.Dependencies, p.DependencyManagement.Dependencies...)
				}
				goto afterImports
			}
			if managedDep.GroupID == groupID && managedDep.ArtifactID == artifactID {
				return managedDep.scope, managedDep.optional, managedDep.Version, parent
			}
		}
	}
	if parent == nil {
		return "", "", "", nil
	}
	return checkParentDependencyManagement(*parent, groupID, artifactID)
}

func figureOutVersion(depwithversion dependency, p pom) string {
	// first check if optional is empty
	if depwithversion.Version == "" {
		// if empty check in dependencyManagement
		if p.DependencyManagement != nil {
			for _, managedDep := range p.DependencyManagement.Dependencies {
				if managedDep.GroupID == depwithversion.GroupID && managedDep.ArtifactID == depwithversion.ArtifactID {
					return figureOutVersion(managedDep, p)
				}
			}
		}
		// if not found check parent's dependencyManagement
		if _, _, version, parent := checkParentDependencyManagement(p, depwithversion.GroupID, depwithversion.ArtifactID); version != "" {
			depwithversion.Version = version
			return figureOutVersion(depwithversion, *parent)
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

		// if not found check for parent by using downloadDepsRepo(currentWorkingRepo,parentGroupID,parentArtifactID,parentVersion)
		value := checkParentProperty(p, propertyName)
		if value != "" {
			return value
		}
		return ""
	}

	return depwithversion.Version
}

func figureOutScope(depwithscope dependency, pom pom) string {
	// first check if optional is empty
	if depwithscope.scope == "" {
		// if empty check in dependencyManagement
		if pom.DependencyManagement != nil {
			for _, managedDep := range pom.DependencyManagement.Dependencies {
				if managedDep.GroupID == depwithscope.GroupID && managedDep.ArtifactID == depwithscope.ArtifactID {
					if managedDep.scope == "" {
						break
					}
					return managedDep.scope
				}
			}
		}

		// if not found check parent's dependencyManagement
		if scope, _, _, p := checkParentDependencyManagement(pom, depwithscope.GroupID, depwithscope.ArtifactID); scope != "" {
			depwithscope.scope = scope
			if scope == "" {
				return ""
			}
			return figureOutScope(depwithscope, *p)
		}
	}

	if strings.HasPrefix(depwithscope.scope, "${") {
		propertyName := strings.TrimSuffix(strings.TrimPrefix(depwithscope.scope, "${"), "}")

		// check for version in pom.Propreties
		if value, ok := pom.Properties[propertyName]; ok {
			if strings.HasPrefix(value, "${") {
				// println(" is this circular ?")
				return checkParentProperty(pom, propertyName)
			}
			return value
		}

		// if not found check for parent by using downloadDepsRepo(currentWorkingRepo,parentGroupID,parentArtifactID,parentVersion)
		value := checkParentProperty(pom, propertyName)
		if value != "" {
			return value
		}
		// if not found return ""
		return ""
	}

	// return optional
	return depwithscope.scope
}

func figureOutOptional(depwithoptional dependency, pom pom) string {
	// first check if optional is empty
	if depwithoptional.optional == "" {
		// if empty check in dependencyManagement
		if pom.DependencyManagement != nil {
			for _, managedDep := range pom.DependencyManagement.Dependencies {
				if managedDep.GroupID == depwithoptional.GroupID && managedDep.ArtifactID == depwithoptional.ArtifactID {
					if managedDep.scope == "" {
						break
					}
					return figureOutOptional(managedDep, pom)
				}
			}
		}

		// if not found check parent's dependencyManagement
		if _, optional, _, p := checkParentDependencyManagement(pom, depwithoptional.GroupID, depwithoptional.ArtifactID); optional != "" {
			depwithoptional.scope = optional
			if optional == "" {
				return ""
			}
			return figureOutScope(depwithoptional, *p)
		}
	}

	if strings.HasPrefix(depwithoptional.optional, "${") {
		propertyName := strings.TrimSuffix(strings.TrimPrefix(depwithoptional.optional, "${"), "}")

		// check for version in pom.Propreties
		if value, ok := pom.Properties[propertyName]; ok {
			if strings.HasPrefix(value, "${") {
				// println(" is this circular ?")
				return checkParentProperty(pom, propertyName)
			}
			return value
		}

		// if not found check for parent by using downloadDepsRepo(currentWorkingRepo,parentGroupID,parentArtifactID,parentVersion)
		value := checkParentProperty(pom, propertyName)
		if value != "" {
			return value
		}
		// if not found return ""
		return ""
	}

	// return optional
	return depwithoptional.optional
}

func figureOutGroupID(depwithgroupid dependency, pom pom) string {
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

		// if not found check for parent by using downloadDepsRepo(currentWorkingRepo,parentGroupID,parentArtifactID,parentVersion)
		if value := checkParentProperty(pom, propertyName); value != "" {
			return value
		}

		// if not found return ""
		return ""
	}

	// if found return groupid
	return depwithgroupid.GroupID
}

func figureOutArtifactID(depwithartifactid dependency, pom pom) string {
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

		// if not found check for parent by using downloadDepsRepo(currentWorkingRepo,parentGroupID,parentArtifactID,parentVersion)
		if value := checkParentProperty(pom, propertyName); value != "" {
			return value
		}

		// if not found return ""
		return ""
	}

	// if found return artifactid
	return depwithartifactid.ArtifactID
}

type Lastest struct {
	Metadata   xml.Name `xml:"metadata"`
	Versioning struct {
		Latest  string `xml:"latest"`
		Release string `xml:"release"`
	} `xml:"versioning"`
}

func figureOutAllLatestAndDownload() {
	for i, s := range latests {
		if i == 0 {
			latests = []string{}
		}
		gas := strings.Split(s, "|")
		groupID := gas[0]
		artifactID := gas[1]
		figureOutLatestAndDownload(groupID, artifactID)
	}
	if len(latests) != 0 {
		figureOutAllLatestAndDownload()
	}

}
func figureOutLatestAndDownload(groupID string, artifactID string) error {
	version, err := figureOutLastest(groupID, artifactID)
	if err != nil {
		return err
	}
	downloadDepsRepo(currentWorkingRepo, groupID, artifactID, version, false)
	return nil
}
func figureOutLastest(groupID string, artifactID string) (string, error) {
	var doc Lastest
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
		return "", fmt.Errorf("failed to download POM: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read POM content: %w", err)
	}
	print("-")
	return string(body), nil
}

type props struct {
	p map[string]string
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
