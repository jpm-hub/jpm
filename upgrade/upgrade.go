package upgrade

import (
	"fmt"
	COM "jpm/common"
	"os"
)

func Upgrade() {
	if COM.IsWindows() {
		tempFile, err := os.CreateTemp("", "upgrade_jpm*.cmd")
		if err != nil {
			return
		}
		tempFile.Close()
		println(tempFile.Name())
		COM.RunCMD(fmt.Sprintf(`curl -L -o "%s" https://cmd.jpmhub.org && start "upgrade" "C:\WINDOWS\system32\cmd.exe" /param="/c %s"`, tempFile.Name(), tempFile.Name()), true)
	} else {
		// Simple background upgrade for Unix
		pid := os.Getpid()
		script := fmt.Sprintf("(while kill -0 %d 2>/dev/null; do sleep 1; done; curl -L -o s.sh https://sh.jpmhub.org && sh s.sh) &", pid)
		COM.RunScript(script, false)
		fmt.Println("Upgrade scheduled to run after process exits.")
	}

}
