package common

import "strings"

func GetJava25AppTemplate() string {
	return `
void main() {
	IO.println("Hello, World");
}`
}

func GetJavaAppTemplate(packaging string, className string) string {
	if packaging == "" {
		packaging = ""
	} else {
		packaging = "package " + packaging + ";"
	}
	return packaging + `
public class ` + className + ` {
	public static void main(String[] args) {
		System.out.println("Hello, World");
	}
	public String run() {
		return "Hello, World";
	}
}`
}
func GetKotlinAppTemplate(packaging string) string {
	if packaging == "" {
		packaging = ""
	} else {
		packaging = "package " + packaging
	}
	return packaging + `
fun main(args: Array<String>) {
	println("Hello, World");
}
fun run():String {
	return "Hello, World";
}`
}

func GetPackageTemplateSimple(language string) string {
	return `main: App` + ifkotlinMain(language) + `
language: ` + language + `
package: App
src: .
scripts: { start: jpm compile && jpm run }
repos: [{ default: https://repo1.maven.org/maven2/ }]
`
}
func GetModuleInfoTemplate(packaging string, lang string) string {
	return `module ` + packaging + ` {
	exports ` + packaging + `;` + ifLangModuleRequires(lang) + `
	requires junit;
}`
}
func ifLangModuleRequires(lang string) string {
	if strings.Contains(lang, "kotlin") {
		return `
		requires kotlin.stdlib;`
	}
	return ""
}
func GetPackageTemplate(packaging string, className string, language string, src string, modular bool) string {
	var p []string
	if packaging == "" {
		p = []string{"app"}
		className = "App"
	} else {
		p = strings.Split(packaging, ".")
		className = strings.ReplaceAll(packaging, "-", "_") + "." + className
	}
	return `main: ` + className + ifkotlinMain(language) + `
version: 0.0.0
language: ` + language + `
package: ` + p[len(p)-1] + ifSrc(src) + ifModular(modular) + `
scripts:
  start : jpm compile && jpm run
  dev: jpm watch _ "jpm start"
  clean: rm -rf out/* dist
dependencies:
repos:
  - default: https://repo1.maven.org/maven2/
`
}

func GetDotVscodeTemplate(src string) string {
	if len(src) == 0 {
		src = "."
	}
	return `{
	"java.project.referencedLibraries": [
		"jpm_dependencies/**/*"
	],
	"java.project.sourcePaths": [
		"` + src + `"
	]
}`
}

func ifSrc(src string) string {
	if len(src) == 0 {
		return ""
	}
	return "\nsrc: " + src
}

func ifModular(modular bool) string {
	if modular {
		return "\nmodular: true"
	}
	return ""
}

func ifkotlinMain(language string) string {
	if language == "kotlin" {
		return "Kt"
	}
	return ""
}

func GetGitignoreTemplate() string {
	return `out
jpm_dependencies/*
!jpm_dependencies/lock.json
dist
*.log
.env
`
}

func GetJavaTestTemplate(packaging string, className string) string {
	iprt := "\nimport " + packaging + "." + className + ";"
	if packaging == "" {
		iprt = ""
	}
	return `package tests;
import org.junit.*;` + iprt + `
public class Test` + className + ` {
	@Test
	public void test() {
		` + className + ` app = new ` + className + `();
		Assert.assertEquals("Hello, World",app.run());
	}
}`
}

func GetKotlinTestTemplate(packaging string, className string) string {
	iprt := "\nimport " + packaging + ".*"
	if packaging == "" {
		iprt = ""
	}
	return `package tests;
import org.junit.*
import kotlin.test.*` + iprt + `
class Test` + className + ` {
	@Test
	fun test() {
		assertEquals("Hello, World",run());
	}
}`
}

func GetDockerFileTempalte(packaging string) string {
	return `FROM eclipse-temurin:25-jdk
WORKDIR /app
COPY dist/ ./
RUN chmod +x ` + packaging + `.sh
CMD ["./` + packaging + `.sh"]
`
}

func GetDockerComposeTempalte(packaging string) string {
	return `version: '3.8'
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: ` + packaging + `
    image: ` + packaging + `:latest
    ports: ["3000:3000"]
`
}

func PrintArt() {
	println(`   ____________  ___
  |_  | ___ \  \/  |  version : ` + VERSION + `
    | | |_/ / .  . |  https://www.jpmhub.org
    | |  __/| |\/| |  The simpler package manager
/\__/ / |   | |  | |  And build tool
\____/\_|   \_|  |_/  for your JVM project`)
}
