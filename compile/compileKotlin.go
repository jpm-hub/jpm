package compile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	COM "jpm/common"
)

func compileKotlin(dir string) error {
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
	if strings.TrimSpace(dir) == "" {
		dir = COM.SrcDir()
	} else {
		_, errS = os.Stat(filepath.Join("jpm_dependencies", "execs"))
		if errS == nil {
			files, err := os.ReadDir(filepath.Join("jpm_dependencies", "execs"))
			if err == nil {
				for _, file := range files {
					if strings.HasSuffix(file.Name(), ".jar") || strings.HasSuffix(file.Name(), ".zip") {
						if builder.Len() > 0 {
							builder.WriteString(separator)
						}
						builder.WriteString("jpm_dependencies/execs/")
						builder.WriteString(file.Name())
					}
				}
			}
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
	if jarFilesString == "" {
		jarFilesString = "."
	}
	var err1 error
	if COM.IsWindows() {
		allKts := findListofAllSrcFile(dir, "*.kt")
		err1 = COM.RunCMD(COM.KOTLINC()+" "+args+" -cp \""+jarFilesString+"\" -d out "+allKts, true)
	} else {
		allKts := findAllSrcFile(dir, "*.kt")
		err1 = COM.RunScript("kotlinc "+args+" -cp \""+jarFilesString+"\" -d out "+allKts, true)
	}

	if err1 != nil {
		fmt.Println("\033[31mkotlin compilation failed\033[0m")
		return fmt.Errorf("compilation failed for kotlin")
	}
	return nil
}
