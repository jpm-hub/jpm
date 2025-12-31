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
var publishing = false
var keepClassifiers = false
var execing = false

func Bundle() {
	COM.FindPackageYML(true)
	packageName = COM.GetSection("package", true).(string)
	// Remove all contents inside "dist" directory, but not the directory itself
	distDir := filepath.Join("dist")
	entries, err := os.ReadDir(distDir)
	if err == nil {
		for _, entry := range entries {
			os.RemoveAll(filepath.Join(distDir, entry.Name()))
		}
	}
	os.RemoveAll(filepath.Join("out"))
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
	classes := ""
	exec := ""
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
			println("Bundling to a fat jar is not yet supported, yet..., help us implement it!")
			os.Exit(1)
		case "-native":
			println("Bundling to native is not yet supported, yet..., help us implement it!")
		case "-exec":
			execing = true
		case "-publish":
			publishing = true
		case "--keep-classifiers":
			keepClassifiers = true
		default:
			println("\t unknown option", arg)
		}
	}
	exec = makeExec(name)
	makePublish(publishing, keepClassifiers)
	println("\033[32mCompiling \033[0m")
	err = COM.RunScript("jpm compile", true)
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

func makeExec(name string) string {
	if !execing {
		return ""
	}
	main := COM.GetSection("main", true).(string)
	if main == "" {
		println("\tpackage.yml must contain main")
		os.Exit(1)
	}
	println("\t --- Bundleling with", main, "as execution Main-Class")
	createScripts(main)
	return "e " + name + ".jar " + main
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
	modular := ""
	if COM.GetSection("modular", false).(bool) {
		modular = "-p jar_libraries"
	}
	args, found := argsMap["java"]
	if !found {
		args = ""
	}
	packageName := COM.GetSection("package", true).(string)
	if publishing {
		if len(modular) > 0 {
			modular = "-p jpm_dependencies:jpm_dependencies/execs"
		}
		modularwin := strings.ReplaceAll(modular, ":", ";")
		depsArg := strings.ReplaceAll(args, "../jpm_dependencies", "jpm_dependencies")
		unix := fmt.Sprintf(`#!/bin/bash
`+COM.ParseEnvVars("export ", true)+`
if $(jpm is-windows > /dev/null ); then
    java %s $(jpm args %s) %s -cp "jpm_dependencies/*;jpm_dependencies/execs/*" %s $@
    exit $?; else
    java %s $(jpm args %s) %s -cp "jpm_dependencies/*:jpm_dependencies/execs/*" %s $@
    exit $?
fi; echo "unknown OS" && exit 1`, depsArg, packageName, modularwin, main, depsArg, packageName, modular, main)
		os.WriteFile(filepath.Join("dist", packageName), []byte(unix+"\n"), 0755)
		return
	}
	depsArg := strings.ReplaceAll(args, "../jpm_dependencies", "jar_libraries")
	unix := fmt.Sprintf("#!/bin/bash\n"+COM.ParseEnvVars("export ", true)+"java %s %s -cp ./*:jar_libraries/* %s $@", depsArg, modular, main)
	depsArg = strings.ReplaceAll(args, "..\\jpm_dependencies", "jar_libraries")
	windows := fmt.Sprintf(COM.ParseEnvVars("set ", false)+"java %s %s -cp ./*;jar_libraries/* %s ", depsArg, modular, main)
	windows = windows + `"%*"`
	os.WriteFile(filepath.Join("dist", packageName), []byte(unix+"\n"), 0755)
	os.WriteFile(filepath.Join("dist", packageName+".cmd"), []byte(windows+"\r\n"), 0755)
}

func makePublish(publishing bool, keepClassifiers bool) {
	if !publishing {
		return
	}
	println("\033[32mInstalling \033[0m")
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
			dependencies.JPM[strings.TrimPrefix(k, classifier)] = v
		}
	}
	for k := range dependencies.JPM {
		if strings.Count(k, "|") == 1 && !strings.HasPrefix(k, "|") {
			delete(dependencies.JPM, k)
		}
	}
	for k, v := range dependencies.Repos {
		for key, value := range v {
			if strings.Count(key, "|") == 2 {
				classifier := strings.Split(key, "|")[0]
				dependencies.Repos[k][strings.TrimPrefix(key, classifier)] = value
			}
		}
		for key := range v {
			if strings.Count(key, "|") == 2 && !strings.HasPrefix(key, "|") {
				delete(dependencies.Repos[k], key)
			}
		}
	}
	return dependencies
}
