//go:build !windows

package upgrade

import (
	COM "jpm/common"
	"os"
	"strings"
)

func upgrade() {

	// if /.sdkman/ is found in the exec path of this, propose to upgrade via sdkman
	execDir, _ := os.Executable()
	if strings.Contains(execDir, "/.sdkman/") {
		println("\033[33mDetected sdkman installation of jpm. To upgrade, run:\n\n  sdk upgrade jpm\n\033[0m")
		return
	}
	// get the directrory of the current executable
	execDir = strings.TrimSuffix(execDir, "/jpm")
	COM.RunScript(`curl -L -o s.sh https://sh.jpmhub.org && sh s.sh `+execDir+`; rm -f s.sh`, true)

}
