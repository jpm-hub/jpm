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
		COM.RunCMD(fmt.Sprintf(`curl -L -o "%s" https://cmd.jpmhub.org && start "upgrade" "C:\WINDOWS\system32\cmd.exe" /param="/c ""%s"""`, tempFile.Name(), tempFile.Name()), true)
	} else {
		// Simple background upgrade for Unix
		COM.RunScript(`curl -L -o s.sh https://sh.jpmhub.org && sh s.sh; rm -f s.sh`, true)
	}

}
