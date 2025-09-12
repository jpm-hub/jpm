package bundle

import (
	"fmt"
	COM "jpm/common"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func Bundle() {

	COM.FindPackageYML()
	packageName := COM.GetSection("package", true).(string)
	os.MkdirAll(filepath.Join("dist", "_dump"), 0755)
	verison := COM.GetSection("version", true).(string)
	if verison == "" {
		println("\tpackage.yml must contain version")
		os.Exit(1)
	}
	name := getName()
	if name == "" {
		name = packageName
		if name == "" {
			println("\tpackage.yml must contain package, or provide a name 'jpm bundle <name>'")
			os.Exit(1)
		}
	}
	packageDir := strings.Split(COM.GetSection("main", true).(string), ".")[0]
	name = name + "-" + verison + ".jar"
	println("\t\033[32mBundleling " + name + " \033[0m")
	main := ""
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
			classes = "-C " + packageDir
			// println("\t --- Bundleling with everything inside of out/ and jpm_depencies/")
			// copyFromOut(filepath.Join("dist", "_dump"))
			// copyFromDependencies(filepath.Join("dist", "_dump"))
			// extractJarAndZip()
			// classes = "-C _dump ."

			// maybe i'll need to implement maven shade here, instead of doing it by hand
			println("Bundling to a fat jar is not yet supported")
			os.Exit(1)
		case "-exe":
			main = COM.GetSection("main", true).(string)
			if main == "" {
				println("\tpackage.yml must contain main")
				os.Exit(1)
			}
			println("\t --- Bundleling with", main, "as execution Main-Class")
			createScripts(main)
			exec = "e " + name + " " + main

		case "-native":

		default:
			println("unknown switch", arg)
		}
	}
	// this means -exe was not triggered
	if exec == "" {
		exec = " " + name
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

	println("\t --- JAR :", filepath.Join("dist", name))
	COM.RunScript("cd dist && "+builder.String(), true)
	if !COM.Verbose {
		os.RemoveAll(filepath.Join("dist", "_dump"))
	}
}
func getName() string {
	if len(os.Args) > 2 && os.Args[2] != "-fat" && os.Args[2] != "-exe" {
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
		COM.RunCMD(cmd, true)
		os.RemoveAll("dist\\_dump\\tests")
	} else {
		cmd := fmt.Sprintf(`rsync -a --exclude 'tests' "%s"/ "%s"/`, src, dst)
		COM.RunScript(cmd, true)
	}
}
func copyFromDependencies(dir string) {
	src := "jpm_dependencies"
	dst := dir
	if COM.IsWindows() {
		cmd := fmt.Sprintf(`xcopy /E /I /Y "%s\" "%s\"`, src, dst)
		COM.RunCMD(cmd, true)
		os.RemoveAll(filepath.Join(dir, ".lock.json"))
		os.RemoveAll(filepath.Join(dir, "tests"))
		os.RemoveAll(filepath.Join(dir, "execs"))
	} else {
		cmd := fmt.Sprintf(`rsync -a --exclude 'tests' --exclude 'execs' "%s"/ "%s"/`, src, dst)
		COM.RunScript(cmd, true)
	}
}
func extractJarAndZip() {}
func createScripts(main string) {
	argsMap := COM.ParseArgs()
	args, found := argsMap["run"]
	if !found {
		args = ""
	}
	unix := fmt.Sprintf(COM.ParseEnvVars("export ")+"java %v -cp ./*:jar_libraries/* %s", args, main)
	windows := fmt.Sprintf(COM.ParseEnvVars("set ")+"java %v -cp ./*;jar_libraries/* %s", args, main)

	os.WriteFile(filepath.Join("dist", "run.sh"), []byte(unix+"\n"), 0755)
	os.WriteFile(filepath.Join("dist", "run.cmd"), []byte(windows+"\r\n"), 0755)
}
