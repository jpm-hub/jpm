package upgrade

import (
	"fmt"
	COM "jpm/common"
	"os"
	"strings"
)

func Upgrade() {
	if COM.IsWindows() {
		tempFile, err := os.CreateTemp("", "upgrade_jpm*.cmd")
		if err != nil {
			return
		}
		tempFile.Close()
		COM.RunCMD(fmt.Sprintf(`curl -L -o "%s" https://cmd.jpmhub.org && start "jpm upgrade" "C:\WINDOWS\system32\cmd.exe" /param="/c ""%s"""`, tempFile.Name(), tempFile.Name()), true)
	} else {
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

}
