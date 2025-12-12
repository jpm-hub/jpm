package init

import (
	"fmt"
	COM "jpm/common"
	"os"
	"strings"
)

func Bare(language string) {
	if err := os.WriteFile("package.yml", []byte(COM.GetPackageTemplateSimple(language)), 0644); err != nil {
		fmt.Printf("Error creating package.yml: %v\n", err)
		os.Exit(1)
	}
	if strings.Contains(language, "kotlin") {
		if err := os.WriteFile("App.kt", []byte(COM.GetKotlinAppTemplate("")), 0644); err != nil {
			fmt.Printf("Error creating main Kotlin file: %v\n", err)
			os.Exit(1)
		}
	}
	if strings.Contains(language, "java") {
		if err := os.WriteFile("App.java", []byte(COM.GetJava25AppTemplate()), 0644); err != nil {
			fmt.Printf("Error creating main Java file: %v\n", err)
			os.Exit(1)
		}
	}
	COM.PrintArt()
	println("\n\tApp is ready to \033[34mrun\033[0m : jpm start")
}
