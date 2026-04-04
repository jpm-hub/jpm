//go:build windows

package upgrade

import (
	"fmt"
	COM "jpm/common"
	"os"
)

func upgrade() {
	tempFile, err := os.CreateTemp("", "upgrade_jpm*.cmd")
	if err != nil {
		return
	}
	tempFile.Close()
	COM.RunCMD(fmt.Sprintf(`curl -L -o "%s" https://cmd.jpmhub.org && start "jpm upgrade" "C:\WINDOWS\system32\cmd.exe" /param="/c ""%s"""`, tempFile.Name(), tempFile.Name()), true)

}
