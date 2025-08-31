package create

import (
	"fmt"
	COM "jpm/common"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v2"
)

var templateYml map[string]any

func Create() {
	if dir, err := os.Getwd(); err == nil {
		ymlPath := filepath.Join(dir, "package.yml")
		if _, err := os.Stat(ymlPath); err == nil {
			fmt.Println("Project already exist int this dir")
			os.Exit(1)
		}
	}

	if len(os.Args) < 3 {
		println("Provide template name")
		os.Exit(1)
	}
	templatePath := ""
	var err error
	template := os.Args[2]
	template = template + ".yml"
	templatesDir := filepath.Join(COM.HomeDir(), "templates")
	if slices.Contains(os.Args, "-yml") {
		templatePath = os.Args[3]
		if !strings.HasSuffix(templatePath, ".yml") {
			templatePath = templatePath + ".yml"
		}
		goto jump
	}

	// Ensure templates directory exists
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		println("Failed to create templates directory:", err.Error())
		os.Exit(1)
	}

	// Download from jpm repo (override local)
	err, _ = COM.DownloadFile(COM.JPM_REPO_API+"templates/"+os.Args[2]+"/"+template, templatesDir, template, true, false)
	if err == nil {
		fmt.Println("\033[32m  template script saved\033[0m")
	}
	// Check if the template file exists
	templatePath = filepath.Join(templatesDir, template)
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		println("Template not found after download:", template)
		os.Exit(1)
	}

jump:
	data, err := os.ReadFile(templatePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := yaml.Unmarshal(data, &templateYml); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//start message
	println(removeSpacer(parse(templateYml["<start-message>"].(string))))

	// run <jpm-init>
	init := parse(templateYml["<jpm-init>"].(string))
	if init != "" {
		args := strings.Split(init, " ")
		script := "jpm init "
		for _, v := range args {
			script = script + sanitize(v) + " "
		}
		if err := COM.RunScript(script, false); err != nil {
			println(err.Error())
			os.Exit(1)
		}
	}
	// run <template-init>
	if val, ok := templateYml["<template-init>"]; ok {
		build(val.([]any))
	}
	// run <templating>
	templating := parse(templateYml["<templating>"].(string))
	if templating != "" {
		if val, ok := templateYml[templating]; ok {
			build(val.([]any))
		}
	}
	// print <finish-message>
	println(removeSpacer(parse(templateYml["<finish-message>"].(string))))
}

func parse(str string) string {
	result := str
	for {
		startIdx := strings.Index(result, "{{")
		if startIdx == -1 {
			break
		}
		endIdx := strings.Index(result[startIdx:], "}}")
		if endIdx == -1 {
			break
		}
		endIdx += startIdx
		varName := strings.TrimSpace(result[startIdx+2 : endIdx])
		if ok, t := verifyIfInput(templateYml[varName]); ok {
			switch t {
			case "text":
				txt := inputText(templateYml[varName].(string))
				templateYml[varName] = txt
				templateYml[varName+".under"] = strings.ReplaceAll(txt, "-", "_")
				templateYml[varName+".title"] = COM.CapitalizeFirst(txt)
				templateYml[varName+".under.title"] = COM.CapitalizeFirst(templateYml[varName+".under"].(string))
				templateYml[varName+".title.under"] = strings.ReplaceAll(templateYml[varName+".title"].(string), "-", "_")
			case "number":
				templateYml[varName] = inputNumber(templateYml[varName].(string))
			case "choice":
				templateYml[varName] = inputChoice(varName, templateYml[varName].(string))
			}
		} else if ok, key, val := verifyIfChoice(varName); ok {
			if prompt, ok := templateYml[key].(string); ok {
				templateYml[key] = inputChoice(key, prompt)
			}
			templateYml[varName] = templateYml[key].(map[string]string)[val]
		} else {
			if _, ok := templateYml[varName]; !ok {
				fmt.Printf("\033[31m%s could not be found \n(do not use .under or .title before getting the value)\033[0m\n", varName)

			}
		}
		val, ok := templateYml[varName].(string)
		if !ok {
			// If not found, remove the unmatched {{KEY}} to avoid infinite loop
			result = result[:startIdx] + result[endIdx+2:]
			continue
		}
		result = result[:startIdx] + val + result[endIdx+2:]
	}
	return result
}

func verifyIfInput(s any) (bool, string) {
	if val, ok := s.(string); ok {
		if strings.Contains(val, "<input-text>") {
			return true, "text"
		}
		if strings.Contains(val, "<input-number>") {
			return true, "number"
		}
		if strings.Contains(val, "<input-choice>") {
			return true, "choice"
		}
	}
	return false, ""
}

func verifyIfChoice(varName string) (bool, string, string) {
	for key := range maps.Keys(templateYml) {
		varSplit := strings.Split(varName, "[")
		prefix := "<choice>:" + varSplit[0] + ":"
		if strings.HasPrefix(key, prefix) {
			return true, varSplit[0], strings.TrimSuffix(varSplit[1], "]")
		}
	}
	return false, "", ""
}

func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			println("invalide jpm init command")
			os.Exit(1)
		}
	}
	return b.String()
}

func build(fileList []any) error {
	for _, v := range fileList {
		if val, ok := v.(map[any]any); ok {
			for k, v := range val {
				key := k.(string)
				value, ok := v.(string)
				if !ok {
					value = ""
				}
				paresedFile := parse(key)
				if COM.IsWindows() {
					paresedFile = strings.ReplaceAll(paresedFile, "/", "\\")
				}
				paresedContent := parse(value)
				if strings.HasSuffix(paresedFile, "/") || strings.HasSuffix(paresedFile, "\\") {
					// it's an empty dir
					os.MkdirAll(paresedFile, 0755)
				} else {
					dir := filepath.Dir(paresedFile)
					if dir != "." {
						if err := os.MkdirAll(dir, 0755); err != nil {
							fmt.Println("Failed to create directory:", err)
							os.Exit(1)
						}
					}
					f, err := os.Create(paresedFile)
					if err != nil {
						fmt.Println("Failed to create file:", err)
						os.Exit(1)
					}
					defer f.Close()
					if _, err := f.WriteString(paresedContent); err != nil {
						fmt.Println("Failed to write to file:", err)
						os.Exit(1)
					}
				}
				break
			}
		} else {
			// os exit, because only the two options of map[string]string and map[string]any are available
		}
	}
	return nil
}

func removeSpacer(str string) string {
	i := 0
	for {
		if strings.HasPrefix(str, "^") {
			str = strings.TrimPrefix(str, "^")
			i += 1
			continue
		}
		str = " " + str
		i -= 1
		if i == 0 {
			break
		}
	}
	return str
}
