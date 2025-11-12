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
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		COM.RunCMD(fmt.Sprintf(`curl -L -o "%s" https://cmd.jpmhub.org && start "%s"`, tempFile.Name(), tempFile.Name()), true)
	} else {
		tempFile, err := os.CreateTemp("", "upgrade_jpm*.sh")
		if err != nil {
			return
		}
		defer os.Remove(tempFile.Name())
		tempFile.Close()
		COM.RunScript(fmt.Sprintf("curl -L -o %s https://sh.jpmhub.org && sh %s", tempFile.Name(), tempFile.Name()), true)
	}

}
