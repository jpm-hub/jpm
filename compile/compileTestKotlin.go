package compile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	COM "jpm/common"
)

func compileTestKotlin() error {
	args := ""
	if allBuildArgs, found := argsMap["kotlinc"]; found {
		args = COM.NormalizeSpaces(allBuildArgs)
	}
	os.MkdirAll(filepath.Join("out", "tests"), 0755)
	_, errS := os.Stat("jpm_dependencies")
	if errS != nil {
		return fmt.Errorf("failed to read ./jpm_dependencies/tests, please run 'jpm install!'")
	}
	jpm_dependenciesFiles, err := os.ReadDir("jpm_dependencies")
	jpm_dependenciesFiles2, err2 := os.ReadDir(filepath.Join("jpm_dependencies", "tests"))
	if err != nil || err2 != nil {
		return fmt.Errorf("failed to read ./jpm_dependencies/tests, please run 'jpm install!'")
	}
	var builder strings.Builder
	for _, file := range jpm_dependenciesFiles2 {
		if strings.HasSuffix(file.Name(), ".jar") {
			if builder.Len() > 0 {
				builder.WriteString(separator)
			}
			builder.WriteString("jpm_dependencies/tests/")
			builder.WriteString(file.Name())
		}
	}
	for _, file := range jpm_dependenciesFiles {
		if strings.HasSuffix(file.Name(), ".jar") || strings.HasSuffix(file.Name(), ".zip") {
			if builder.Len() > 0 {
				builder.WriteString(separator)
			}
			builder.WriteString("jpm_dependencies/")
			builder.WriteString(file.Name())
		}
	}
	jarFilesString := builder.String()

	var err4 error
	if COM.IsWindows() {
		allkts := strings.Join([]string{findAllSrcFile(COM.SrcDir(), "."), findAllSrcFile("tests", ".")}, " ")
		err4 = COM.RunCMD(COM.KOTLINC()+" "+args+" -cp \""+"out;"+jarFilesString+"\" -d out "+allkts, true)
	} else {
		allkts := strings.Join([]string{findAllSrcFile(COM.SrcDir(), "*.kt"), findAllSrcFile("tests", "*.kt")}, " ")
		err4 = COM.RunScript("kotlinc  -cp \""+"out:"+jarFilesString+"\" "+args+" -d out "+allkts, true)
	}
	if err4 != nil {
		return fmt.Errorf("\033[31m test compilation failed for kotlin\033[0m")
	}
	return nil
}
