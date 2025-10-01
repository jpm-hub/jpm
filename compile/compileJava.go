package compile

import (
	"fmt"
	"os"
	"strings"

	COM "jpm/common"
)

func compileJava() error {
	args := ""
	if allBuildArgs, found := argsMap["javac"]; found {
		args = allBuildArgs
	}
	jpm_dependenciesFiles, err := os.ReadDir("jpm_dependencies")
	if err != nil {
		return fmt.Errorf("failed to read jpm_dependencies directory: %s", err.Error())
	}
	var builder strings.Builder
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
	if jarFilesString == "" {
		jarFilesString = " "
	}
	r, w, originalOut, err := startCheckLastLineForErrors()
	if err != nil {
		return err
	}
	allJavas := findAllSrcFile(COM.SrcDir(), "*.java")
	var err1 error
	if COM.IsWindows() {
		err1 = COM.RunCMD(COM.JAVAC()+" "+args+" -p jpm_dependencies -cp \""+jarFilesString+"\" -d out "+allJavas, true)
	} else {
		err1 = COM.RunScript(COM.JAVAC()+" "+args+" -p jpm_dependencies -cp \""+jarFilesString+"\" -d out "+allJavas, true)
	}
	err = endCheckLastLineForErrors(r, w, originalOut)
	if err1 != nil || err != nil {
		fmt.Println("\033[31mjava compilation failed\033[0m")
		return fmt.Errorf("compilation failed for java")
	}
	return nil
}
