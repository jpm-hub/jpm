package init

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	COM "jpm/common"
)

func Init(cliargs []string) {
	if dir, err := os.Getwd(); err == nil {
		ymlPath := filepath.Join(dir, "package.yml")
		if _, err := os.Stat(ymlPath); err == nil {
			fmt.Println("Project already initialized")
			os.Exit(1)
		}
	}

	projectName := "app"
	diffName := false
	language := "java"
	handled := []string{}
	git := false
	docker := false
	if len(cliargs) >= 2 {
		for i, arg := range cliargs {
			if i < 2 {
				continue
			}
			if slices.Contains(handled, arg) {
				continue
			}
			handled = append(handled, arg)
			if arg == "-git" {
				git = true
				continue
			}
			if arg == "-docker" {
				docker = true
				continue
			}
			if arg == "-kt" {
				language = "kotlin"
				continue
			}
			if arg == "-java" {
				language = "java"
				continue
			}

			if arg == "-mod" {
				println("new module")
				continue
			}
			if strings.Contains(arg, "\\") {
				println("use foward slashes instead")
				os.Exit(1)
			}
			diffName = true
			projectName = arg
			projectNameSlice := strings.Split(projectName, ".")
			if strings.Contains(projectName, ".") {
				className := COM.CapitalizeFirst(projectNameSlice[len(projectNameSlice)-1])
				projectName = strings.Join(projectNameSlice[:len(projectNameSlice)-1], ".") + "." + className
				projectName = strings.Trim(projectName, ".")
			}
		}
		if !diffName {
			fmt.Println("Initializing project App")
		}
	}

	appPathSlice := strings.Split(strings.ReplaceAll(projectName, ".", "/"), "/")
	className := COM.CapitalizeFirst(appPathSlice[len(appPathSlice)-1]) // App
	appMainJavaFile := className + ".java"                              // App.java
	appMainKotlinFile := className + ".kt"                              // App.kt
	appDir := appPathSlice[0]
	if len(appPathSlice) > 1 {
		appDir = strings.Join(appPathSlice[:len(appPathSlice)-1], "/") // com/app
	}

	p := strings.Split(appDir, "/")
	appDir = strings.ReplaceAll(appDir, "-", "_")
	packaging := p[len(p)-1]
	count := strings.Count(projectName, "/")
	src := ""
	if count > 0 {
		pSlice := strings.Split(projectName, "/")
		p := pSlice[len(pSlice)-1]
		pSlice = strings.Split(p, ".")
		packaging = strings.Join(pSlice[:len(pSlice)-1], ".")
		if packaging == "" {
			packaging = appPathSlice[len(appPathSlice)-1]
			appDir = filepath.Join(appDir, strings.ReplaceAll(packaging, "-", "_"))
		}
		src = strings.SplitN(projectName, ".", 2)[0]
		srcSlice := strings.Split(src, "/")
		src = strings.Join(srcSlice[:len(srcSlice)-1], "/")
	}
	appMainJavaFile = appDir + "/" + appMainJavaFile     // com/app/App.java
	appMainKotlinFile = appDir + "/" + appMainKotlinFile // com/app/App.kt

	// Create directories
	if err := os.MkdirAll(appDir, 0755); err != nil {
		fmt.Printf("Error creating %v directory: %v\n", appDir, err)
		os.Exit(1)
	}
	if err := os.MkdirAll("tests", 0755); err != nil {
		fmt.Printf("Error creating tests directory: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Join("jpm_dependencies", "tests"), 0755); err != nil {
		fmt.Printf("Error creating jpm_dependencies/tests directory: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(filepath.Join(".vscode"), 0755); err == nil {
		os.WriteFile(filepath.Join(".vscode", "settings.json"), []byte(COM.GetDotVscodeTemplate(src)), 0644)
	}
	switch language {
	case "java":
		initJava(appMainJavaFile, packaging, className, src)
	case "kotlin":
		initKotlin(appMainKotlinFile, packaging, className, src)
	}
	COM.CopyToDependencies(language)
	if git {
		COM.RunScript("git init", true)
	}
	if docker {
		if err := os.WriteFile("Dockerfile", []byte(COM.GetDockerFileTempalte()), 0644); err != nil {
			fmt.Printf("Error creating package.yml: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile("docker-compose.yml", []byte(COM.GetDockerComposeTempalte(packaging)), 0644); err != nil {
			fmt.Printf("Error creating test file: %v\n", err)
			os.Exit(1)
		}
	}
	if len(cliargs) == 2 {
		COM.PrintArt()
		println("\n\tyour new project is ready to \033[34mrun\033[0m : jpm start")
		println("\tyour new project is ready to \033[32mtest\033[0m : jpm test \n")
	}
}
