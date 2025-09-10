package compile

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	COM "jpm/common"
)

var separator string = ":"
var argsMap map[string]string

func initCompile() {
	COM.FindPackageYML()
	os.MkdirAll("out", 0755)
	os.MkdirAll(filepath.Join("jpm_dependencies", "tests"), 0755)
	if COM.IsWindows() {
		separator = ";"
	}
	argsMap = COM.ParseArgs()
}
func Compile() error {
	initCompile()
	languageList := findLanguages()
	var err error
	var err2 error = nil
	for _, v := range languageList {
		switch v {
		case "java":
			err = compileJava()
		case "kotlin":
			err = compileKotlin()
		default:
			err = compileJava()
		}
		if err != nil {
			println(err.Error())
			err2 = fmt.Errorf("%s", err.Error())
		}
	}
	return err2
}

func CompileTest() error {
	initCompile()
	languageList := findLanguages()
	var err error
	var err2 error = nil
	for _, v := range languageList {
		switch v {
		case "java":
			err = compileTestJava()
		case "kotlin":
			err = compileTestKotlin()
		default:
			err = compileTestJava()
		}
		if err != nil {
			println(err.Error())
			err2 = fmt.Errorf("%s", err.Error())
		}
	}
	return err2
}

func findLanguages() []string {
	languageList := []string{}
	languages_any := COM.GetSection("language", false)
	if languages, ok := languages_any.(string); ok {
		if strings.Contains(languages, ",") {
			languageList = strings.Split(languages, ",")
			for i, v := range languageList {
				languageList[i] = COM.NormalizeSpaces(v)
			}
		} else {
			languageList = append(languageList, COM.NormalizeSpaces(languages))
		}
	} else {
		languageList = append(languageList, "java")
	}
	return languageList
}

func startCheckLastLineForErrors() (read *os.File, write *os.File, originalStdout *os.File, err error) {
	r, w, err := os.Pipe()
	if err != nil {
		println("failed to capture stdout")
		return nil, nil, nil, err
	}
	originalStdout = os.Stdout
	os.Stdout = w
	return r, w, originalStdout, nil
}
func endCheckLastLineForErrors(r *os.File, w *os.File, originalStdout *os.File) error {
	w.Close()
	os.Stdout = originalStdout
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	if err != nil {
		return err
	}
	lines := strings.Split(buf.String(), "\n")
	if len(lines) == 0 {
		return nil // No lines to check
	}
	println(lines[len(lines)-1])
	if strings.Contains(lines[len(lines)-1], "errors") || strings.HasPrefix(lines[len(lines)-1], "error:") {
		return errors.New("last line contains 'errors'")
	}
	return nil
}

func findAllSrcFile(dir string, fileWildcard string) string {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		println(" The package name should be a dir in the root directory :", dir)
		os.Exit(1)
		return ""
	}
	// Recursively add all subdirectories, but only include if a file matches the wildcard in the dir
	str := strings.Builder{}
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			matches, _ := filepath.Glob(filepath.Join(path, fileWildcard))
			if len(matches) > 0 {
				str.WriteString(filepath.Join(path, fileWildcard))
				str.WriteString(" ")
			}
		}
		return nil
	})
	return str.String()
}
