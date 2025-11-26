package compile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	COM "jpm/common"
)

func compileJava(dir string) error {
	args := ""
	if allBuildArgs, found := argsMap["javac"]; found {
		args = allBuildArgs
	}

	jpm_dependenciesFiles := []os.DirEntry{}
	var builder strings.Builder
	_, errS := os.Stat("jpm_dependencies")
	if errS == nil {
		var err error
		jpm_dependenciesFiles, err = os.ReadDir("jpm_dependencies")
		if err != nil {
			jpm_dependenciesFiles = []os.DirEntry{}
		}
	}
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
		jarFilesString = " "
	}
	r, w, originalOut, err := startCheckLastLineForErrors()
	if err != nil {
		return err
	}

	allJavas := findAllSrcFile(dir, "*.java")
	var err1 error
	if COM.IsWindows() {
		err1 = COM.RunCMD("javac "+args+" -p jpm_dependencies -cp \""+jarFilesString+"\" -d out "+allJavas, true)
	} else {
		err1 = COM.RunScript("javac "+args+" -p jpm_dependencies -cp \""+jarFilesString+"\" -d out "+allJavas, true)
	}
	err = endCheckLastLineForErrors(r, w, originalOut)
	if err1 != nil || err != nil {
		fmt.Println("\033[31mjava compilation failed\033[0m")
		return fmt.Errorf("compilation failed for java")
	}
	return nil
}
