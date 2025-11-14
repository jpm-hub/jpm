package compile

import (
	"fmt"
	"os"
	"strings"

	COM "jpm/common"
)

func compileKotlin() error {
	args := ""
	if allBuildArgs, found := argsMap["kotlinc"]; found {
		args = COM.NormalizeSpaces(allBuildArgs)
	}
	jpm_dependenciesFiles := []os.DirEntry{}
	_, errS := os.Stat("jpm_dependencies")
	if errS == nil {
		var err error
		jpm_dependenciesFiles, err = os.ReadDir("jpm_dependencies")
		if err != nil {
			jpm_dependenciesFiles = []os.DirEntry{}
		}
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
		jarFilesString = "."
	}

	var err1 error
	if COM.IsWindows() {
		allKts := findListofAllSrcFile(COM.SrcDir(), "*.kt")
		err1 = COM.RunCMD(COM.KOTLINC()+" "+args+" -cp \""+jarFilesString+"\" -d out "+allKts, true)
	} else {
		allKts := findAllSrcFile(COM.SrcDir(), "*.kt")
		err1 = COM.RunScript("kotlinc "+args+" -cp \""+jarFilesString+"\" -d out "+allKts, true)
	}

	if err1 != nil {
		fmt.Println("\033[31mkotlin compilation failed\033[0m")
		return fmt.Errorf("compilation failed for kotlin")
	}
	return nil
}
