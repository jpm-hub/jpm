package install

import (
	"errors"
	"fmt"
	COM "jpm/common"
	"strings"
)

func makeGHJsonFileName(username, pack, ver string) string {
	return fmt.Sprintf("gh.%s.%s-%s.json", username, pack, ver)
}

func generateGithubDepUrl(username, pack, ver, filename string) string {
	return fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", username, pack, ver, filename)
}

func disectGithubDepString(depString string) (jpmDependency, error) {
	dSlice := strings.Split(depString, " ")
	if len(dSlice) < 3 {
		println("\033[31m" + tab + "The dependency : " + depString + " is ambigious and will be ignored\033[0m")
		return jpmDependency{}, errors.New("The dependency : " + depString + " is ambigious")
	}
	artver := strings.Split(dSlice[2], ":")
	t := ""
	if len(dSlice) == 3 {
		t = ""
	} else if len(dSlice) == 4 {
		t = dSlice[3]
	} else {
		println("\033[31m" + tab + "The dependency : " + depString + " is ambigious\033[0m")
		return jpmDependency{}, errors.New("The dependency : " + depString + " is ambigious")
	}
	version := ""
	if len(artver) == 2 {
		version = artver[1]
	}
	return jpmDependency{
		Alias:      dSlice[0],
		Type:       "github",
		GhUsername: dSlice[1],
		Package:    artver[0],
		Version:    COM.NormalizeSpaces(version),
		Scope:      t,
	}, nil
}

func saveAllGithubSubDependencies(d *jpmDependency, alias string) {
	version := ""

	switch d.Version {
	case "latest":
		// figure out version for this github dependency
		v, err := figureOutLatestGithub(d)
		if err != nil {
			println("\033[31m  --- " + alias + ": Resolving " + d.Package + " ! " + "Unable to get latest version\033[0m\n")
			return
		}
		version = v
	case "":
		// figure out version for this github dependency
		v, err := figureOutLatestGithub(d)
		if err != nil {
			println("\033[31m  --- " + alias + ": Resolving " + d.Package + " ! " + "Unable to get latest version\033[0m\n")
			return
		}
		version = v
		if alias != "default" {
			COM.ReplaceDependency(fmt.Sprintf("%s %s %s %s", alias, d.GhUsername, d.Package, d.Scope), fmt.Sprintf("%s %s %s:%s %s", alias, d.GhUsername, d.Package, d.Version, d.Scope))
		} else {
			COM.ReplaceDependency(fmt.Sprintf("%s %s %s", d.GhUsername, d.Package, d.Scope), fmt.Sprintf("%s %s:%s %s", d.GhUsername, d.Package, d.Version, d.Scope))
		}
	default:
		version = d.Version
	}
	d.Version = version
	al := ">"
	if alias != "default" && alias != ">" {
		al = strings.ToUpper(alias) + ":"
	}
	println("\033[32m  --- " + al + " Resolving " + d.Package + ":" + version + "\033[0m")
	print("      Resolving   [")
	saveJPMExecToDownloadList(d)
	url := generateGithubDepUrl(d.GhUsername, d.Package, d.Version, "dependencies.json")
	downloadDepsJPM(d, url)
	println("]")
}
func addGithubSubDependenciesToDownloadList() {

}
func figureOutLatestGithub(d *jpmDependency) (string, error) {

	return "latest", nil
}
