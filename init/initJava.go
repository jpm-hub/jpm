package init

import (
	"fmt"
	COM "jpm/common"
	"os"
	"strings"
)

func initJava(appMainJavaFile string, packaging string, className string, src string) {
	// Write package.yml
	if err := os.WriteFile("package.yml", []byte(COM.GetPackageTemplate(packaging, className, "java", src)), 0644); err != nil {
		fmt.Printf("Error creating package.yml: %v\n", err)
		os.Exit(1)
	}
	// Write test file
	if err := os.WriteFile("tests/Test"+className+".java", []byte(COM.GetJavaTestTemplate(strings.ReplaceAll(packaging, "-", "_"), className)), 0644); err != nil {
		fmt.Printf("Error creating test file: %v\n", err)
		os.Exit(1)
	}

	// Write .gitignore
	if err := os.WriteFile(".gitignore", []byte(COM.GetGitignoreTemplate()), 0644); err != nil {
		fmt.Printf("Error creating .gitignore: %v\n", err)
		os.Exit(1)
	}

	// Write main Java file
	if err := os.WriteFile(appMainJavaFile, []byte(COM.GetJavaAppTemplate(strings.ReplaceAll(packaging, "-", "_"), className)), 0644); err != nil {
		fmt.Printf("Error creating main Java file: %v\n", err)
		os.Exit(1)
	}

}
