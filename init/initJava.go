package init

import (
	"fmt"
	COM "jpm/common"
	"os"
	"strings"
)

func initJava(appMainJavaFile string, packaging string, className string, lang string, src string, modular bool) {
	// Write package.yml
	if err := os.WriteFile("package.yml", []byte(COM.GetPackageTemplate(packaging, className, lang, src, modular)), 0644); err != nil {
		fmt.Printf("Error creating package.yml: %v\n", err)
		os.Exit(1)
	}
	if packaging != "" {
		// Write test file
		if err := os.WriteFile("tests/Test"+className+".java", []byte(COM.GetJavaTestTemplate(strings.ReplaceAll(packaging, "-", "_"), className)), 0644); err != nil {
			fmt.Printf("Error creating test file: %v\n", err)
			os.Exit(1)
		}
		// Write main Java file
		if err := os.WriteFile(appMainJavaFile, []byte(COM.GetJavaAppTemplate(strings.ReplaceAll(packaging, "-", "_"), className)), 0644); err != nil {
			fmt.Printf("Error creating main Java file: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := os.WriteFile(appMainJavaFile, []byte(COM.GetJava25AppTemplate()), 0644); err != nil {
			fmt.Printf("Error creating main Java file: %v\n", err)
			os.Exit(1)
		}
	}
}
