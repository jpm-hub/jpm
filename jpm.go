package main

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

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
		fmt.Println(" \033[33mjpm init\033[0m  <project> [-git, -docker] :\n\t\tInitializes a project default: App")
		fmt.Println("\t\t-git : will iniitalize a git repo")
		fmt.Println("\t\t-docker : will create a dockerfile and a docker-compose.yml")
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
		fmt.Println(" \033[33mjpm watch\033[0m [\"(src/**)\"] [\"<command>\"] :")
		fmt.Println("\t\tRuns a command on filec change")
		fmt.Println("\t\t\"(src/**)\": determines the file that will trigger the command ( _ to ignore)")
		fmt.Println("\t\t\"<command>\": execute this command before compilation when files change ( _ to ignore)")
		println()
		fmt.Println(" \033[33mjpm bundle\033[0m [name] [-fat, -exec, -publish, --keep-classifiers] :") // add windows ach, arm ach and linux and darwin and all one day maybe
		fmt.Println("\t\tCreates a jar")
		fmt.Println("\t\t-fat: Creates a fat jar (all inccluded, one file)")
		fmt.Println("\t\t-exec: Creates an executable jar and scripts")
		fmt.Println("\t\t-publish: creates necessary files to publish")
		fmt.Println("\t\t--keep-classifiers: Keeps the classifiers in the dependencies")
		println()
		fmt.Println(" \033[33mjpm test\033[0m :\tRuns tests with junit under tests/")
		println()
		fmt.Println(" \033[33mjpm install\033[0m [-f] :")
		fmt.Println("\t\tInstalls the dependencies from package.yml")
		fmt.Println("\t\t-f: force re-install ")
		println()
		fmt.Println(" \033[33mjpm install -repo\033[0m <alias>:<url> :")
		fmt.Println("\t\tAdds a repo to package.yml")
		fmt.Println("\t\texample -> 'jpm install -repo mvn:https://repo1.maven.org/maven2/'")
		println()
		fmt.Println(" \033[33mjpm install <package>:<version> <scope> \033[0m \t\t\tdependency: a package from jpm repository")
		fmt.Println(" \033[33mjpm install <alias> <GroupID> <ArtifactID>:<version> <scope> \033[0m\tdependency: from a repo in package.yml")
		fmt.Println(" \033[33mjpm install raw [-x] <link-to-your-file> <scope> \033[0m\t\tdependency: from url")
		fmt.Println("\t\t<alias>: An alias defined in pakcage.yml ")
		fmt.Println("\t\t<scope>: may be ommited if dependecy is for compilation/runtime")
		fmt.Println("\t\t<scope>: may be test if dependecy is for tests")
		fmt.Println("\t\t<scope>: may be exec if dependecy is a utility")
		fmt.Println("\t\t-x: extract zip or tar.gz files from raw dependencies")
		return
	case "doctor":
		if !DOC.Doctor(false, true) {
			os.Exit(1)
		}
		return
	case "setup":
		sw := []string{"-java", "-kotlin", "-junit", "-hotswap", "-v", "-jpx"}
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
		if COMPILE.Compile() != nil {
			os.Exit(1)
		}
	case "run":
		execOverride("run")
		if RUN.Run() != nil {
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
			println("\n" + err.Error())
			os.Exit(1)
		}
	case "i":
		fallthrough
	case "install":
		execOverride("install")
		INSTALL.Install()
	default:
		argsStr := ""
		if len(os.Args) > 1 {
			argsStr = strings.TrimSpace(strings.Join(os.Args[2:], " "))
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
		println("\033[33mOmitting Override: "+"'"+"jpm", sc+"'", "\033[0m")
		return
	}
	if os.Getenv("JPM_OVERRIDE") != sc {
		// Join all the args from os.Args except os.Args[0] into a string
		argsStr := ""
		if len(os.Args) > 1 {
			argsStr = strings.TrimSpace(strings.Join(os.Args[2:], " "))
		}
		if slices.Contains(scriptsS, sc+"@") {
			println("\033[33mOverriding: "+"'"+"jpm", sc+"'", "for "+"'"+"jpm", sc+"@"+"'", "\033[0m")
			cmd := "export JPM_OVERRIDE=" + sc + "\nexport PATH=\"$PATH:$(pwd)/jpm_dependencies/execs\"\n" + strings.ReplaceAll(scripts[sc+"@"], "...args@", argsStr)
			if err := COM.RunScript(cmd, true); err != nil {
				fmt.Printf("Error running script '%s': %v \n", sc, err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}
}
