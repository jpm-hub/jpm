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
			return COM.RunCMD(COM.ParseEnvVars(prefix)+"cd out && "+COM.JAVA()+" "+args+" -p ../jpm_dependencies -cp \".;./*;../jpm_dependencies/*\" "+mainClass, true)
		} else {
			return COM.RunScript(COM.ParseEnvVars(prefix)+"cd out && "+COM.JAVA()+" "+args+" -p ../jpm_dependencies -cp \".:./*:../jpm_dependencies/*\" "+mainClass, true)
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
			if COM.IsWindows() {
				return COM.RunCMD(COM.ParseEnvVars(prefix)+"cd out && "+COM.JAVA()+" "+args+" -p ../jpm_dependencies -XX:HotswapAgent=fatjar -XXaltjvm=dcevm -javaagent:"+COM.HomeDir()+"\\libs\\hotswap-agent.jar=autoHotswap=true,"+argshs+" -cp \".;./*;../jpm_dependencies/*\" "+mainClass+" "+runArgs, true)
			} else {
				return COM.RunScript(COM.ParseEnvVars(prefix)+"cd out && "+COM.JAVA()+" "+args+" -p ../jpm_dependencies -XX:HotswapAgent=fatjar -XXaltjvm=dcevm -javaagent:"+COM.HomeDir()+"/libs/hotswap-agent.jar=autoHotswap=true,"+argshs+" -cp \".:./*:../jpm_dependencies/*\" "+mainClass+" "+runArgs, true)
			}
		} else {
			for i := 2; i < len(os.Args); i++ {
				if i > 2 {
					runArgs += " "
				}
				runArgs += os.Args[i]
			}
			if COM.IsWindows() {
				return COM.RunCMD(COM.ParseEnvVars(prefix)+"cd out && "+COM.JAVA()+" "+args+" -p ../jpm_dependencies -cp \".;./*;../jpm_dependencies/*\" "+mainClass+" "+runArgs, true)
			} else {
				return COM.RunScript(COM.ParseEnvVars(prefix)+"cd out && "+COM.JAVA()+" "+args+" -p ../jpm_dependencies -cp \".:./*:../jpm_dependencies/*\" "+mainClass+" "+runArgs, true)
			}
		}
	}
	println("run command has wrong args")
	return nil
}
