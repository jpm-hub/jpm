package install

import (
	"fmt"
	COM "jpm/common"
	"os"
	"slices"
	"strings"

	"github.com/juju/errors"
)

func installFromCLI(path string) {
	aliases := findExistingAliases(path)
	depString := strings.Join(os.Args[2:], " ")
	depString = COM.NormalizeSpaces(depString)
	aliases = append(aliases, "raw")
	var err error
	if !slices.Contains(aliases, strings.ToLower(strings.Split(depString, " ")[0])) {
		// is from jpm
		depString, err = fromJPMCLI(depString)
	} else if strings.Split(depString, " ")[0] != "raw" {
		// is from repo
		depString, err = fromRepoCLI(path, depString)

	} else {
		// is from raw
		err = fromRawCLI(depString)
	}
	if err == nil {
		COM.AddToSection("dependencies", depString)
	}

}

func findExistingAliases(path string) []string {
	repos := getRepoList()
	aliases := []string{}
	for _, v := range repos.Repos {
		aliases = append(aliases, v.Alias)
	}
	return aliases
}

func fromJPMCLI(depString string) (string, error) {
	// load in the json
	println("JPM cli Instalation is not yet implemented")
	return "", errors.New("not implemented")
}

func fromRepoCLI(path string, depString string) (string, error) {
	// load in the json
	loadDependencies()
	repoList := getRepoList()
	alias := strings.ToLower(strings.Split(depString, " ")[0])
	for _, v := range repoList.Repos {
		if v.Alias == alias {
			dep, err := disectRepoDepString(COM.NormalizeSpaces(depString), v.Repo, alias)
			if err != nil {
				break
			}
			currentOuterScope = dep.DependencyType
			currentWorkingRepo = dep.Repo
			err = saveAllRepoSubDependencies(&dep)
			depString := fmt.Sprintf("%s %s %s:%s %s", dep.Alias, dep.GroupID, dep.ArtefactID, dep.ArtVer, dep.DependencyType)
			depString = strings.TrimSpace(depString)
			if err == nil {
				addRepoSubDependenciesToDownloadList(dep)
				return depString, nil
			}
			return "", err
		}
	}
	println("Could not find " + alias + " in your repos section in package.yml")
	return "", nil
}

func fromRawCLI(depString string) error {
	// load in the json
	return nil
}
