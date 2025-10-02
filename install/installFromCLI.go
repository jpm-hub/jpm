package install

import (
	"fmt"
	COM "jpm/common"
	"os"
	"strings"
)

func installFromCLI() {
	aliases := findExistingAliases()
	depString := strings.Join(os.Args[2:], " ")
	depString = strings.ReplaceAll(depString, "--save-dev", "exec")
	depString = COM.NormalizeSpaces(depString)
	aliases = append(aliases, "raw")
	jpmDeps, noneJpmdeps := findAllJPM([]string{depString}, aliases)
	var err error
	if len(jpmDeps) > 0 {
		depString, err = fromJPMCLI(depString)
	} else if len(noneJpmdeps) > 0 {
		// is from repo
		depString, err = fromRepoCLI(depString)
	} else {
		// is from raw
		depString, err = fromRawCLI(depString)
	}
	if err == nil {
		COM.AddToSection("dependencies", depString)
	} else {
		os.Exit(1)
	}
}
func fromJPMCLI(depString string) (string, error) {
	originalDepString := depString
	loadLockDependencies()
	depString = fromJPM([]string{depString})
	installDependencies()
	dumpDependencies()
	cleanup()
	if depString == "" {
		return "", fmt.Errorf("could not resolve %s", originalDepString)
	}
	s := strings.Split(originalDepString, ":")
	if len(s) == 2 && strings.Contains(s[1], "latest") {
		return COM.NormalizeSpaces(originalDepString), nil
	}
	return depString, nil
}

func fromRepoCLI(depString string) (string, error) {
	loadLockDependencies()
	repoList := getRepoList()
	sdep := strings.SplitN(depString, " ", 2)
	alias := strings.ToLower(sdep[0])
	for _, v := range repoList.Repos {
		if v.Alias == "default" && !strings.Contains(sdep[1], "/") {
			alias = "default"
			depString = "default " + depString
		}

		if v.Alias == alias {
			dep, err := disectRepoDepString(COM.NormalizeSpaces(depString), v.Repo, alias)
			if err != nil {
				break
			}
			currentOuterScope = dep.Scope
			currentWorkingRepo = dep.Repo
			err = saveAllRepoSubDependencies(&dep)
			depString := ""
			if alias == "default" {
				depString = fmt.Sprintf("%s %s:%s %s", dep.GroupID, dep.ArtefactID, dep.ArtVer, dep.Scope)
			} else {
				depString = fmt.Sprintf("%s %s %s:%s %s", dep.Alias, dep.GroupID, dep.ArtefactID, dep.ArtVer, dep.Scope)
			}
			depString = strings.TrimSpace(depString)
			if err == nil {
				addRepoSubDependenciesToDownloadList(dep.Repo)
				installDependencies()
				dumpDependencies()
				cleanup()
				return depString, nil
			}
			return "", err
		}
	}
	println("Could not find " + alias + " in your repos section in package.yml")
	return "", nil
}

func fromRawCLI(depString string) (string, error) {
	fromRAW([]string{COM.NormalizeSpaces(depString)})
	return COM.NormalizeSpaces(depString), nil
}
