package bundle

import (
	COM "jpm/common"
	//INSTALL "jpm/install"
	"os"
)

func executeMultipleBundles(packages []string, topName string, topVersion string) {
	println("Bundling multiple packages is not yet supported")
	os.Exit(1)
	classified := COM.GetSection("classified", false).(bool)
	for _, pack := range packages {

		COM.RunScript("cp dist/"+topName+"-"+topVersion+".jar "+pack+"/jpm_dependencies/"+topName+"-"+topVersion+".jar", true)
		err := os.Chdir(pack)
		if err != nil {
			println("\tCould not bundle package " + pack)
			os.Exit(1)
		}
		COM.FindPackageYML(true)
		packName := COM.GetSection("package", true).(string)
		if packName != pack {
			println("\t", "Did not find package "+pack, "in "+pack+"/package.yml")
			os.Exit(1)
		}
		Bundle()
		if !classified {
			COM.RunScript("mv dist/*.jar ../dist/", true)
		} else {

		}
		err = os.Chdir("..")
		if err != nil {
			println("\tCould not bundle package " + pack)
			os.Exit(1)
		}
	}
}
