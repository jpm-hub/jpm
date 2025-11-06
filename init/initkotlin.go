package init

import (
	"fmt"
	COM "jpm/common"
	"os"
	"strings"
)

func initKotlin(appMainKotlinFile string, packaging string, className string, lang string, src string) {
	// Write package.yml
	if err := os.WriteFile("package.yml", []byte(COM.GetPackageTemplate(packaging, className, lang, src)), 0644); err != nil {
		fmt.Printf("Error creating package.yml: %v\n", err)
		os.Exit(1)
	}
	if packaging != "" {
		// Write test file
		if err := os.WriteFile("tests/Test"+className+".kt", []byte(COM.GetKotlinTestTemplate(strings.ReplaceAll(packaging, "-", "_"), className)), 0644); err != nil {
			fmt.Printf("Error creating test file: %v\n", err)
			os.Exit(1)
		}
	}

	// Write main Java file
	if err := os.WriteFile(appMainKotlinFile, []byte(COM.GetKotlinAppTemplate(strings.ReplaceAll(packaging, "-", "_"))), 0644); err != nil {
		fmt.Printf("Error creating main Java file: %v\n", err)
		os.Exit(1)
	}
}
