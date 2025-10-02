package bundle

import (
	"encoding/json"
	"fmt"
	COM "jpm/common"
	INSTALL "jpm/install"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var packageName string
var version string

func Bundle() {
	COM.FindPackageYML(true)
	packageName = COM.GetSection("package", true).(string)
	os.RemoveAll(filepath.Join("dist"))
	os.MkdirAll(filepath.Join("dist", "_dump"), 0755)
	version = COM.GetSection("version", true).(string)
	if version == "" {
		println("\tpackage.yml must contain version")
		os.Exit(1)
	}
	name := packageName + "-" + version
	if n := getName(); n != "" {
		name = n
	}
	println("\033[32mBundleling " + name + ".jar \033[0m")
	main := ""
	classes := ""
	exec := ""
	publishing := false
	keepClassifiers := false
	var builder strings.Builder
	builder.WriteString("jar cf")
	handled := []string{}
	for i, arg := range os.Args {
		if i < 2 {
			continue
		}
		if slices.Contains(handled, arg) {
			continue
		}
		handled = append(handled, arg)
		switch arg {
		case "-fat":
			println("Bundling to a fat jar is not yet supported")
			os.Exit(1)
		case "-native":
			println("Bundling to native is not yet supported, yet...")
		case "-exec":
			main = COM.GetSection("main", true).(string)
			if main == "" {
				println("\tpackage.yml must contain main")
				os.Exit(1)
			}
			println("\t --- Bundleling with", main, "as execution Main-Class")
			createScripts(main)
			exec = "e " + name + ".jar " + main
		case "-publish":
			publishing = true
		case "--keep-classifiers":
			keepClassifiers = true
		default:
			println("\t unknown option", arg)
		}
	}
	println("\033[32mInstalling \033[0m")
	makePublish(publishing, keepClassifiers)
	println("\033[32mCompiling \033[0m")
	err := COM.RunScript("jpm compile", true)
	if err != nil {
		os.Exit(1)
	}
	// this means -exec was not triggered
	if exec == "" {
		exec = " " + name + ".jar"
	}
	builder.WriteString(exec)

	// this means -fat was not triggered (only compiled classes will be in jar)
	if classes == "" {
		classes = " -C _dump ."
		os.MkdirAll(filepath.Join("dist", "jar_libraries"), 0755)
		copyFromOut()
		copyFromDependencies(filepath.Join("dist", "jar_libraries"))
	}
	builder.WriteString(classes)

	println("\t --- JAR :", filepath.Join("dist", name+".jar"))
	COM.RunScript("cd dist && "+builder.String(), true)
	packs := COM.GetSection("packages", false).([]string)
	if len(packs) > 0 {
		executeMultipleBundles(packs, name, version)
	}
	if !COM.Verbose {
		os.RemoveAll(filepath.Join("dist", "_dump"))
	}
}

func getName() string {
	if len(os.Args) > 2 && os.Args[2] != "-fat" && os.Args[2] != "-exec" && os.Args[2] != "-native" && os.Args[2] != "-publish" {
		// classified := COM.GetSection("classified", false).(bool)
		// if classified {
		// 	println("\033[31musing a custom name with classified=true is not recommended\033[0m")
		// }
		return os.Args[2]
	}
	return ""
}

func copyFromOut() {
	src := "out"
	dst := "dist/_dump"
	destwin := "dist\\_dump"
	if COM.IsWindows() {
		cmd := fmt.Sprintf(`xcopy /E /I /Y "%s\" "%s\"`, src, destwin)
		COM.RunCMD(cmd, false)
		os.RemoveAll("dist\\_dump\\tests")
	} else {
		cmd := fmt.Sprintf(`rsync -a --exclude 'tests' "%s"/ "%s"/`, src, dst)
		COM.RunScript(cmd, false)
	}
}
func copyFromDependencies(dir string) {
	src := "jpm_dependencies"
	dst := dir
	if COM.IsWindows() {
		cmd := fmt.Sprintf(`xcopy /E /I /Y "%s\" "%s\"`, src, dst)
		COM.RunCMD(cmd, false)
		os.RemoveAll(filepath.Join(dir, "lock.json"))
		os.RemoveAll(filepath.Join(dir, "tests"))
		os.RemoveAll(filepath.Join(dir, "execs"))
	} else {
		cmd := fmt.Sprintf(`rsync -a --exclude 'lock.json' --exclude 'tests' --exclude 'execs' "%s"/ "%s"/`, src, dst)
		COM.RunScript(cmd, false)
	}
}
func extractJarAndZip() {}
func createScripts(main string) {
	argsMap := COM.ParseArgs()
	args, found := argsMap["java"]
	if !found {
		args = ""
	} else {
		args = " " + args
	}
	unixArgs := strings.ReplaceAll(args, "../jpm_dependencies", "jar_libraries")
	unix := fmt.Sprintf("#!/bin/bash\n"+COM.ParseEnvVars("export ")+"java%s -p jar_libraries -cp ./*:jar_libraries/* %s $@", unixArgs, main)
	winArgs := strings.ReplaceAll(args, "..\\jpm_dependencies", "jar_libraries")
	windows := fmt.Sprintf(COM.ParseEnvVars("set ")+"java%s -p jar_libraries -cp ./*;jar_libraries/* %s ", winArgs, main)
	windows = windows + `"%*"`
	os.WriteFile(filepath.Join("dist", COM.GetSection("package", true).(string)), []byte(unix+"\n"), 0755)
	os.WriteFile(filepath.Join("dist", COM.GetSection("package", true).(string)+".cmd"), []byte(windows+"\r\n"), 0755)
}

func makePublish(publishing bool, keepClassifiers bool) {
	if !publishing {
		return
	}
	depsJson := removeClassifiers(removeScopes(INSTALL.QuickInstall(publishing)), keepClassifiers)
	fp := filepath.Join("dist", "dependencies.json")

	file, err := os.Create(fp)
	if err != nil {
		fmt.Println("Error creating dependencies.json:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(depsJson); err != nil {
		fmt.Println("Error encoding dependencies.json:", err)
	}
	// create README.md file with package name version and description from package.yml
	pack := COM.GetSection("package", true).(string)
	version := COM.GetSection("version", true).(string)
	description := COM.GetSection("description", true).(string)
	if description == "" {
		description = "No description provided."
	}
	readmeContent := fmt.Sprintf("# %s\n\nVersion: %s\n\nDescription: %s\n", pack, version, description)
	os.WriteFile(filepath.Join("dist", "README.md"), []byte(readmeContent), 0644)
}

func removeScopes(dependencies COM.Dependencies) COM.Dependencies {
	for k, v := range dependencies.JPM {
		dependencies.JPM[strings.TrimSuffix(k, "|")] = v
		if strings.HasSuffix(k, "|") {
			delete(dependencies.JPM, k)
		}
	}
	for k, v := range dependencies.Repos {
		for key, value := range v {
			dependencies.Repos[k][strings.TrimSuffix(key, "|")] = value
			if strings.HasSuffix(key, "|") {
				delete(dependencies.Repos[k], key)
			}
		}
	}
	maps.DeleteFunc(dependencies.JPM, func(key string, v string) bool {
		if strings.HasSuffix(key, "|test") || strings.HasSuffix(key, "|exec") {
			return true
		}
		return false
	})
	for k, _ := range dependencies.Repos {
		maps.DeleteFunc(dependencies.Repos[k], func(key string, v string) bool {
			if strings.HasSuffix(key, "|test") || strings.HasSuffix(key, "|exec") {
				return true
			}
			return false
		})
	}
	return dependencies
}

func removeClassifiers(dependencies COM.Dependencies, keepClassifiers bool) COM.Dependencies {
	if keepClassifiers {
		return dependencies
	}
	for k, v := range dependencies.JPM {
		if strings.Count(k, "|") == 1 {
			classifier := strings.Split(k, "|")[0]
			dependencies.JPM[strings.TrimPrefix(k, classifier+"|")] = v
			delete(dependencies.JPM, k)
		}
	}
	for k, v := range dependencies.Repos {
		for key, value := range v {
			if strings.Count(key, "|") == 2 {
				classifier := strings.Split(key, "|")[0]
				dependencies.Repos[k][strings.TrimPrefix(key, classifier+"|")] = value
				delete(dependencies.Repos[k], key)
			}
		}
	}
	return dependencies
}
