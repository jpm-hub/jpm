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
	jpm_dependenciesFiles, err := os.ReadDir("jpm_dependencies")
	jpm_dependenciesFiles2, err2 := os.ReadDir(filepath.Join("jpm_dependencies", "tests"))
	if err != nil || err2 != nil {
		return fmt.Errorf("failed to read ./jpm_dependencies/tests")
	}
	var builder strings.Builder
	for _, file := range jpm_dependenciesFiles2 {
		if strings.HasSuffix(file.Name(), ".jar") {
			if builder.Len() > 0 {
				builder.WriteString(separator)
			}
			builder.WriteString(filepath.Join("jpm_dependencies", "tests") + slash)
			builder.WriteString(file.Name())
		}
	}
	for _, file := range jpm_dependenciesFiles {
		if strings.HasSuffix(file.Name(), ".jar") || strings.HasSuffix(file.Name(), ".class") || strings.HasSuffix(file.Name(), ".zip") {
			if builder.Len() > 0 {
				builder.WriteString(separator)
			}
			builder.WriteString("jpm_dependencies" + slash)
			builder.WriteString(file.Name())
		}
	}
	jarFilesString := builder.String()

	allkts := strings.Join([]string{findAllSrcFile(COM.SrcDir(), "*.kt"), findAllSrcFile("tests", "*.kt")}, " ")
	// err3 := COM.RunScript(COM.KOTLINC()+" -include-runtime -cp \""+jarFilesString+"\" -d out/tests "+allkts, true)
	// allkts =
	err4 := COM.RunScript(COM.KOTLINC()+" -cp \""+"out/tests:"+jarFilesString+"\" "+args+" -d out/tests "+allkts, true)
	if err4 != nil {
		return fmt.Errorf("\033[31m test compilation failed for kotlin\033[0m")
	}
	return nil
}
