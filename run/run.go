package run

import (
	COM "jpm/common"
	COMPILE "jpm/compile"
	WATCH "jpm/watch"
	"os"
)

func Run() error {
	COM.FindPackageYML(true)
	mainClass := COM.GetSection("main", true).(string)
	if mainClass == "" {
		println("no main set in package.yml")
		return nil
	}
	argsMap := COM.ParseArgs()
	args, found := argsMap["java"]
	if !found {
		args = ""
	}
	prefix := "export "
	if COM.IsWindows() {
		prefix = "set "
	}
	if len(os.Args) == 2 {
		if COM.IsWindows() {
			return COM.RunCMD(COM.ParseEnvVars(prefix, false)+"cd out && "+COM.JAVA()+" "+args+" -p ../jpm_dependencies -cp \".;./*;../jpm_dependencies/*\" "+mainClass, true)
		} else {
			return COM.RunScript(COM.ParseEnvVars(prefix, true)+"cd out && "+COM.JAVA()+" "+args+" -p ../jpm_dependencies -cp \".:./*:../jpm_dependencies/*\" "+mainClass, true)
		}
	}
	allRunArgs, found := argsMap["hotswap"]
	argshs := ""
	if found {
		argshs = allRunArgs
	}
	if len(os.Args) > 2 {
		runArgs := ""
		if os.Args[2] == "-hot" {
			if len(os.Args) > 5 {
				for i := 5; i < len(os.Args); i++ {
					if i > 5 {
						runArgs += " "
					}
					runArgs += os.Args[i]
				}
			}
			COMPILE.Compile()
			go WATCH.Watch(true)
			createHotswapConfig()
			defer os.Remove("out/hotswap-agent.properties")
			if COM.IsWindows() {
				return COM.RunCMD(COM.ParseEnvVars(prefix, false)+"cd out && "+COM.JAVA()+" "+args+" -p ../jpm_dependencies -XX:HotswapAgent=fatjar -XXaltjvm=dcevm -javaagent:"+COM.HomeDir()+"\\libs\\hotswap-agent.jar=autoHotswap=true,"+argshs+" -cp \".;./*;../jpm_dependencies/*\" "+mainClass+" "+runArgs, true)
			} else {
				return COM.RunScript(COM.ParseEnvVars(prefix, true)+"cd out && "+COM.JAVA()+" "+args+" -p ../jpm_dependencies -XX:HotswapAgent=fatjar -XXaltjvm=dcevm -javaagent:"+COM.HomeDir()+"/libs/hotswap-agent.jar=autoHotswap=true,"+argshs+" -cp \".:./*:../jpm_dependencies/*\" "+mainClass+" "+runArgs, true)
			}

		} else {
			for i := 2; i < len(os.Args); i++ {
				if i > 2 {
					runArgs += " "
				}
				runArgs += os.Args[i]
			}
			if COM.IsWindows() {
				return COM.RunCMD(COM.ParseEnvVars(prefix, false)+"cd out && "+COM.JAVA()+" "+args+" -p ../jpm_dependencies -cp \".;./*;../jpm_dependencies/*\" "+mainClass+" "+runArgs, true)
			} else {
				return COM.RunScript(COM.ParseEnvVars(prefix, true)+"cd out && "+COM.JAVA()+" "+args+" -p ../jpm_dependencies -cp \".:./*:../jpm_dependencies/*\" "+mainClass+" "+runArgs, true)
			}
		}
	}
	println("run command has wrong args")
	return nil
}

func createHotswapConfig() {
	// stat out/hotswap-agent.properties
	_, err := os.Stat("hotswap-agent.properties")
	if err == nil {
		COM.CopyFile("hotswap-agent.properties", "out/hotswap-agent.properties")
		return
	}
	// write file with content
	content := `
autoHotswap=true
LOGGER=RELOAD
extraClasspath=
watchResources=
webappDir=
excludedClassLoaderPatterns=
	`
	err = os.WriteFile("hotswap-agent.properties", []byte(content), 0644)
	os.WriteFile("out/hotswap-agent.properties", []byte(content), 0644)
	if err != nil {
		println("Error creating hotswap-agent.properties:", err.Error())
	}
}
