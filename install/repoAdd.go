package install

import (
	COM "jpm/common"
	"os"
	"strings"
)

func addNewRepo(repoString string) {
	aliasAndaddr := strings.SplitN(repoString, ":", 2)
	alias := strings.ToLower(aliasAndaddr[0])
	addr := aliasAndaddr[1]
	if !strings.HasSuffix(addr, "/") {
		addr = addr + "/"
	}
	if strings.HasPrefix(alias, ">") || strings.HasPrefix(alias, "-") || alias == "raw" || alias == "jpm" || !strings.HasPrefix(addr, "http") {
		println("\t", repoString, "must not have - before it, must not be raw or jpm, must have http:// or https://\n")
		os.Exit(1)
	}
	repoSection := COM.GetSection("repos", false)
	if repoSection == nil {
		COM.AddToSection("repos", map[string]string{alias: addr})
		return
	}
	for _, v := range repoSection.([]any) {
		if _, ok := v.(map[string]string)[alias]; ok {
			println("alias :" + alias + " already existant, please override it manually in package.yml")
			return
		}
	}
	COM.AddToSection("repos", map[string]string{alias: addr})
	addFinishMessage("\033[33mInfo : added repo " + alias + " : " + addr + " \033[0m")
}
