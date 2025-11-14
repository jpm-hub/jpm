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
	files, err2 := os.ReadDir(".")
	if err2 != nil {
		fmt.Println("Failed to read current directory:", err2)
		os.Exit(1)
	}
	if len(files) > 0 {
		fmt.Print("\033[38;5;208mDirectory is not empty. Continue? (y/N): \033[0m")
		var response string
		fmt.Scanf("%s", &response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Aborted.")
			os.Exit(0)
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
		fmt.Println("\033[32mtemplate script saved\033[0m")
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
	println(removeSpacer(templateYml["<start-message>"].(string)))

	// run <init-script>
	if err := COM.RunScript(parse(templateYml["<init-script>"].(string)), true); err != nil {
		println(err)
		os.Exit(1)
	}
	// run <template-init>
	if val, ok := templateYml["<template-init>"]; ok {
		build(val.([]any))
	}
	// run <templating>
	templating := parse(templateYml["<templating>"])
	if templating != "" {
		if val, ok := templateYml[templating]; ok {
			build(val.([]any))
		}
	}
	// print <finish-message>
	println(removeSpacer(parse(templateYml["<finish-message>"])))
}

func parse(str any) string {
	result, ok := str.(string)
	if !ok {
		return ""
	}
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
		if _, ok := templateYml[varName]; ok {
			if _, ok := templateYml[varName+"is_jpm_defined"]; ok {
				result = result[:startIdx] + templateYml[varName].(string) + result[endIdx+2:]
				continue
			}
			if txt, ok := templateYml[varName].(string); ok {
				addTexModifiers(txt, varName)
				templateYml[varName+"is_jpm_defined"] = "true"
			}
		}
		if ok, t := verifyIfInput(templateYml[varName]); ok {
			switch t {
			case "text-sanitized":
				txt := inputTextSanitized(templateYml[varName].(string))
				addTexModifiers(txt, varName)
			case "text":
				txt := inputText(templateYml[varName].(string))
				addTexModifiers(txt, varName)
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
		} else if ok, val := verifyIfConditionnal(varName); ok {
			templateYml[varName] = val
		} else {
			if _, ok := templateYml[varName]; !ok {
				fmt.Printf("\033[31m%s could not be found \n(do not refer to .snake or .camel before getting the value)\033[0m\n", varName)

			}
		}
		templateYml[varName+"is_jpm_defined"] = "true"
		val, ok := templateYml[varName].(string)
		if !ok {
			// If not found, remove the unmatched {{KEY}} to avoid infinite loop
			result = result[:startIdx] + result[endIdx+2:]
			continue
		}
		result = result[:startIdx] + val + result[endIdx+2:]
	}
	result = strings.ReplaceAll(result, "{\\{", "{{")
	result = strings.ReplaceAll(result, "}\\}", "}}")
	return result
}

func addTexModifiers(txt, varName string) {
	templateYml[varName] = txt
	templateYml[varName+".snake"] = strings.ReplaceAll(txt, "-", "_")
	templateYml[varName+".camel"] = COM.CapitalizeFirst(txt)
}

func verifyIfInput(s any) (bool, string) {
	if val, ok := s.(string); ok {
		if strings.Contains(val, "<input-text-sanitized>") {
			return true, "text-sanitized"
		}
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

func verifyIfConditionnal(varName string) (bool, string) {
	// {{if-eq lang[language] kotlin -> - mvn org.jetbrains.kotlin kotlin-test:latest test}}
	if strings.HasPrefix(varName, "if-eq ") {
		varName = COM.NormalizeSpaces(varName)
		vars := strings.SplitN(varName, " ", 5)
		if templateYml[vars[1]] == vars[2] && vars[3] == "->" && len(vars[4]) > 0 {
			return true, vars[4]
		}
		return true, ""
	}
	return false, ""
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
				paresedContent := parse(value)
				if strings.HasSuffix(paresedFile, "/") {
					// it's an empty dir
					os.MkdirAll(paresedFile, 0755)
				} else if strings.HasPrefix(paresedFile, "!") {
					// remove this file or dir
					pathToRemove := strings.TrimPrefix(paresedFile, "!")
					os.RemoveAll(pathToRemove)
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
	if !strings.HasPrefix(str, "^") {
		return str
	}
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
