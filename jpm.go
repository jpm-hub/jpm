package main

import (
	"fmt"
	"os"
	"slices"

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
// add support for jpm install (from a jpm repo) (no derived dependencies)
// add jpm bundle -fat (creates a fat jar)
// add docker to init
// add windows support (scripts and back slashes)
func main() {
	COM.Init()
	if len(os.Args) == 1 {
		COM.PrintArt()
		println()
		DOC.Doctor(false)
		return
	}
	scriptName := os.Args[1]

	switch scriptName {
	case "-h":
		fallthrough
	case "--h":
		fallthrough
	case "-help":
		COM.PrintArt()
		fmt.Println("\n\033[33m jpm doctor\033[0m :\tCheck JPM's issues")
		println()
		fmt.Println(" \033[33mjpm init\033[0m  <project_name> [-git, -docker] :\n\t\tInitializes a project default: App")
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
		fmt.Println(" \033[33mjpm bundle\033[0m [name] [-fat, -exe] :") // add windows ach, arm ach and linux and darwin and all one day maybe
		fmt.Println("\t\tCreates a jar")
		fmt.Println("\t\t-fat: Creates a fat jar (all inccluded, one file)")
		fmt.Println("\t\t-exe: Creates an executable jar and scripts")
		fmt.Println(" \033[33mjpm test\033[0m :\tRuns tests with junit under tests/")
		println()
		fmt.Println(" \033[33mjpm install\033[0m [-f] :\tInstalls the dependencies from package.yml")
		fmt.Println("\t\t-f: force re-install ")
		println()
		fmt.Println(" \033[33mjpm install -repo\033[0m <alias>:<url> :\tAdds a new repository to package.yml")
		fmt.Println("\t\texample -> 'jpm install -repo mvn:https://repo1.maven.org/maven2/'")
		println()
		fmt.Println(" \033[33mjpm install <package>:<version> <scope> \033[0m \t\t\tdependency: a package from jpm repository")
		fmt.Println(" \033[33mjpm install <repo> <GroupID> <ArtifactID>:<version> <scope> \033[0m\tdependency: from a repo in package.yml")
		fmt.Println(" \033[33mjpm install raw [-x] <link-to-your-file> <scope> \033[0m\t\tdependency: from url")
		fmt.Println("\t\t<repo>: An alias defined in pakcage.yml ")
		fmt.Println("\t\t<scope>: may be ommited if dependecy is for compilation/runtime")
		fmt.Println("\t\t<scope>: may be test if dependecy is for tests")
		fmt.Println("\t\t<scope>: may be exec if dependecy is a utility")
		fmt.Println("\t\t-x: extract zip or tar.gz files from raw dependencies")
		return
	case "doctor":
		DOC.Doctor(false)
	case "init":
		INIT.Init()
	case "create":
		CREATE.Create()
	case "compile":
		if COMPILE.Compile() != nil {
			os.Exit(1)
		}
	case "run":
		if RUN.Run() != nil {
			os.Exit(1)
		}
	case "watch":
		WATCH.Watch(false)
	case "bundle":
		BUNDLE.Bundle()
	case "test":
		if err := TEST.TestScript(); err != nil {
			println("\n" + err.Error())
			os.Exit(1)
		}
	case "i":
		fallthrough
	case "install":
		INSTALL.Install()
	case "setup":
		if len(os.Args) == 3 {
			sw := []string{"-java", "-kotlin", "-junit", "-HotSwapAgent", "-verbose"}
			if slices.Contains(sw, os.Args[2]) {
				SETUP.Setup(os.Args[2])
				os.Exit(0)
			}
		}
		fallthrough
	default:
		COM.FindPackageYML()
		scripts := COM.ParseScripts()
		scriptCmd, found := scripts[scriptName]
		if !found {
			fmt.Printf("Script '%s' not found in package.yml\n", scriptName)
			os.Exit(1)
		}
		cmd := scriptCmd
		if err := COM.RunScript(cmd, true); err != nil {
			fmt.Printf("Error running script '%s': %v \n", scriptName, err)
			os.Exit(1)
		}
	}

}
