package install

import (
	"fmt"
	COM "jpm/common"
)

func makeGHJsonFileName(username, pack, ver string) string {
	return fmt.Sprintf("gh.%s.%s-%s.json", username, pack, ver)
}

func generateGithubDepUrl(username, pack, ver, filename string) string {
	return fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", username, pack, ver, filename)
}

func disectGithubDepString(d string) (jpmDependency, error) {

	return jpmDependency{}, nil
}

func saveAllGithubSubDependencies(d *jpmDependency, alias string) {
	version := ""
	v, err := figureOutLatestGithub(d)

	switch d.Version {
	case "latest":
		// figure out version for this github dependency
		if err != nil {
			println("\033[31m  --- " + alias + ": Resolving " + d.Package + " ! " + "Unable to get latest version\033[0m\n")
			return
		}
		version = v
	case "":
		// figure out version for this github dependency
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
	println("\033[32m  --- " + alias + ": Resolving " + d.Package + ":" + version + "\033[0m")
	print("      Resolving   [")
	saveGithubExecToDownloadList(d)
	downloadDepsJPM(d)
	println("]")
}

func saveGithubExecToDownloadList(d *jpmDependency) {

}

func figureOutLatestGithub(d *jpmDependency) (string, error) {

	return "latest", nil
}
