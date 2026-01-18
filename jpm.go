package main

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	ARGS "jpm/args"
	BUNDLE "jpm/bundle"
	COM "jpm/common"
	COMPILE "jpm/compile"
	CREATE "jpm/create"
	DOC "jpm/doctor"
	INIT "jpm/init"
	INSTALL "jpm/install"
	RUN "jpm/run"
	SETUP "jpm/setup"
	TEST "jpm/test_script"
	UPGRADE "jpm/upgrade"
	WATCH "jpm/watch"
)

// TODO:
// add comment support for package.yml (when updating repos and deps from cli)
// add jpm bundle -fat (creates a fat jar)
// add jpm bundle -native (creates a native executable with graalvm maybe or jlink)
// add suport for additionnal classpath and modulepath in package.yml
// add detection of version downgrade in install
// add support for inner projects and local dependencies
// add support for multiple packages compilation and bundling to create classified jars
// add -with for classiified installs
// add -no for excluding dependencies
var scriptsS []string = []string{}
var scripts map[string]string

func main() {
	COM.Init()
	if len(os.Args) == 1 {
		COM.PrintArt()
		println()
		DOC.Doctor(false, false)
		return
	}
	scriptName := os.Args[1]
	switch scriptName {
	case "-v":
		fallthrough
	case "--v":
		fallthrough
	case "-version":
		fallthrough
	case "--version":
		println("version :", COM.VERSION)
		return
	case "-h":
		fallthrough
	case "--h":
		fallthrough
	case "-help":
		fallthrough
	case "--help":
		COM.PrintArt()
		fmt.Println("\n\033[33m jpm doctor\033[0m :\tCheck JPM's issues")
		println()
		fmt.Println("\033[33m jpm upgrade\033[0m :\tupgrades JPM to the latest version")
		println()
		fmt.Println(" \033[33mjpm init\033[0m  <project> [-git, -docker, -kt, -java, -mod, -add] :\n\t\tInitializes a project default: App")
		fmt.Println("\t\t-git : will iniitalize a git repo")
		fmt.Println("\t\t-docker : will create a dockerfile and a docker-compose.yml")
		fmt.Println("\t\t-kt : will iniitalize a kotlin project")
		fmt.Println("\t\t-java : will iniitalize a java project")
		fmt.Println("\t\t-mod : will iniitalize a project with module-info.java")
		fmt.Println("\t\t-add : will iniitalize a new sub project")
		println()
		fmt.Println(" \033[33mjpm create\033[0m [-yml] <jpm_template> :")
		fmt.Println("\t\tCreates a project from a template")
		fmt.Println("\t\t-yml: will look for the <jpm_template>.yml in the working dir")
		println()
		fmt.Println(" \033[33mjpm <script>\033[0m :\truns a script from package.yml")
		println()
		fmt.Println(" \033[33mjpm compile\033[0m :\tCompiles the project")
		println()
		fmt.Println(" \033[33mjpm run\033[0m [-hot] [\"(src/**)\"] [\"<command>\"] [...<app_args>] :")
		fmt.Println("\t\tRuns the class files in out/")
		fmt.Println("\t\t-hot: will hot reload classes in memomy on file change")
		fmt.Println("\t\t\"(src/**)\": determines the file that will trigger change ( _ to ignore)")
		fmt.Println("\t\t\"<command>\": execute this command before compilation when files change ( _ to ignore)")
		fmt.Println("\t\t\"<app_args>\": arguments for the app ")
		println()
		fmt.Println(" \033[33mjpm watch\033[0m [-r] [\"(src/**)\"] [\"<command>\"] :")
		fmt.Println("\t\tRuns a command on file change")
		fmt.Println("\t\t\"-r\": kills the previous command ran, before running it again")
		fmt.Println("\t\t\"(src/**)\": determines the file that will trigger the command ( _ to ignore)")
		fmt.Println("\t\t\"<command>\": execute this command before compilation when files change ( _ to ignore)")
		println()
		fmt.Println(" \033[33mjpm bundle\033[0m [name] [-fat, -exec, -publish, --keep-classifiers] :") // add windows ach, arm ach and linux and darwin and all one day maybe
		fmt.Println("\t\tCreates a jar")
		fmt.Println("\t\t-fat: Creates a fat jar (all inccluded, one file)")
		fmt.Println("\t\t-exec: Creates an executable jar and scripts")
		fmt.Println("\t\t-publish: creates necessary files to publish")
		fmt.Println("\t\t--keep-classifiers: Keeps the classifiers in the dependencies.json")
		println()
		fmt.Println(" \033[33mjpm test\033[0m [junit_args...] : Runs tests with junit under tests/")
		println()
		fmt.Println(" \033[33mjpm install \033[0m [-f, -update, -with <classifier>, -no <package>, -repo <alias>:<url>] <dependency>,... :")
		fmt.Println("\t\tInstalls the dependencies from package.yml, dependency list seperated by commas")
		fmt.Println("\t\t-f: force re-install ")
		fmt.Println("\t\t-update: updates all dependencies to latest versions")
		fmt.Println("\t\t-with <classifier>: adds a classifier to all dependencies")
		fmt.Println("\t\t-with <groupID>:<classifier> : adds a classifier to all dependencies with the given groupID")
		fmt.Println("\t\t-with <package>:<classifier> : adds a classifier to the given dependency")
		fmt.Println("\t\t-no   <package> : excludes a package from installation")
		fmt.Println("\t\t-repo <alias>:<url> : Adds a repo with an alias to package.yml")
		println()
		fmt.Println(" \033[33m\t -> <package>:<version> <scope>,... \033[0m \t\t\t\t<dependency> : a package from jpm repository")
		fmt.Println(" \033[33m\t -> <alias> <GroupID> <ArtifactID>:<version> <scope>,... \033[0m\t<dependency> : from a repo in package.yml")
		fmt.Println(" \033[33m\t -> raw [-x] <link-to-your-file> <scope>,... \033[0m\t\t\t<dependency> : from url")
		fmt.Println("\t\t<version>: may be a specific version, or ommited to get latest, or 'latest'")
		fmt.Println("\t\t<alias>: An alias defined in pakcage.yml, may be ommited if alias is 'default'")
		fmt.Println("\t\t<scope>: may be ommited, or 'test' for testing purposes, or 'exec' for utility purposes")
		fmt.Println("\t\t-x: extract zip or tar.gz files from raw dependencies")
		return
	case "doctor":
		if !DOC.Doctor(false, true) {
			os.Exit(1)
		}
		return
	case "is-windows":
		if COM.IsWindows() {
			os.Exit(0)
		}
		os.Exit(1)
	case "args":
		if len(os.Args) > 2 {
			os.Stdout.Write([]byte(ARGS.Args(os.Args[2])))
			return
		}
		os.Exit(1)
		return
	case "upgrade":
		UPGRADE.Upgrade()
		return
	case "setup":
		sw := []string{"-java", "-kotlin", "-junit", "-hotswap", "-v", "-jpx", "-clear"}
		if len(os.Args) > 2 && slices.Contains(sw, os.Args[2]) {
			SETUP.Setup(os.Args[2])
			return
		} else {
			fmt.Println("Available options are:", sw)
			os.Exit(1)
		}
	}

	COM.FindPackageYML(false)
	scripts = COM.ParseScripts()
	scriptCmd, found := scripts[scriptName]
	for k := range maps.Keys(scripts) {
		scriptsS = append(scriptsS, k)
	}

	scriptName = strings.TrimSuffix(scriptName, "!")
	switch scriptName {
	case "init":
		execOverride("init")
		INIT.Init(os.Args)
	case "create":
		execOverride("create")
		CREATE.Create()
	case "compile":
		execOverride("compile")
		args := []string{}
		if len(os.Args) > 2 {
			args = os.Args[2:]
		}
		if COMPILE.Compile(args...) != nil {
			os.Exit(1)
		}
	case "run":
		execOverride("run")
		err := RUN.Run()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	case "watch":
		execOverride("watch")
		WATCH.Watch(false)
	case "bundle":
		execOverride("bundle")
		BUNDLE.Bundle()
	case "test":
		execOverride("test")
		if err := TEST.TestScript(); err != nil {
			println("\n Tests failed")
			os.Exit(1)
		}
	case "ci":
		execOverride("ci")
		COM.RunScript("rm -rf ./jpm_dependencies/*", false)
		fallthrough
	case "i":
		fallthrough
	case "install":
		execOverride("install")
		INSTALL.Install()
	default:
		argsStr := ""
		if len(os.Args) > 1 {
			if COM.IsWindows() {
				argsStr = strings.TrimSpace(strings.Join(os.Args[2:], "' '"))
				argsStr = "'" + argsStr + "'"
				println(argsStr)
			} else {
				argsStr = strings.TrimSpace(strings.Join(os.Args[2:], " "))
			}
		}
		if !found {
			fmt.Printf("Script '%s' not found in package.yml\n", scriptName)
			os.Exit(1)
		}
		cmd := "export PATH=\"$PATH:$(pwd)/jpm_dependencies/execs\"\n" + strings.ReplaceAll(scriptCmd, "...args@", argsStr)
		if err := COM.RunScript(cmd, true); err != nil {
			fmt.Printf("Error running script '%s': %v \n", scriptName, err)
			os.Exit(1)
		}
	}

}

func execOverride(sc string) {
	if strings.HasSuffix(os.Args[1], "!") {
		if COM.Verbose {
			println("\033[33mOmitting Override: "+"'"+"jpm", sc+"'", "\033[0m")
		}
		return
	}
	if os.Getenv("JPM_OVERRIDE") != sc {
		// Join all the args from os.Args except os.Args[0] into a string
		argsStr := ""
		if len(os.Args) > 1 {
			argsStr = strings.TrimSpace(strings.Join(os.Args[2:], " "))
		}
		if slices.Contains(scriptsS, sc+"@") {
			if COM.Verbose {
				println("\033[33mOverriding: "+"'"+"jpm", sc+"'", "for "+"'"+"jpm", sc+"@"+"'", "\033[0m")
			}
			cmd := "export JPM_OVERRIDE=" + sc + "\nexport PATH=\"$PATH:$(pwd)/jpm_dependencies/execs\"\n" + strings.ReplaceAll(scripts[sc+"@"], "...args@", argsStr)
			if err := COM.RunScript(cmd, true); err != nil {
				fmt.Printf("Error running script '%s': %v \n", sc, err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}
}
