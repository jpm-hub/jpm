package common

import "strings"

func GetJavaAppTemplate(packaging string, className string) string {
	return `package ` + packaging + `;
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
	return `package ` + packaging + `;
fun main(args: Array<String>) {
	println("Hello, World");
}
fun run():String {
	return "Hello, World";
}`
}

func GetPackageTemplate(packaging string, className string, language string, src string) string {
	p := strings.Split(packaging, ".")
	return `main: ` + strings.ReplaceAll(packaging, "-", "_") + "." + className + ifkotlinMain(language) + `
version: 0.0.0
language: ` + language + `
package: ` + p[len(p)-1] + ifSrc(src) + `
scripts:
  start : jpm compile && jpm run
  dev: jpm watch _ "jpm start"
  clean: rm -rf out/*
dependencies:` + ifkotlinTest(language) + `
repos:
  - mvn: https://repo1.maven.org/maven2/
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
		"` + src + `",
		"tests"
	]
}`
}

func ifSrc(src string) string {
	if len(src) == 0 {
		return ""
	} else {
		return "\nsrc: " + src
	}
}
func ifkotlinTest(language string) string {
	if language == "kotlin" {
		return `
  - mvn org.jetbrains.kotlin kotlin-test:latest test
  - mvn org.jetbrains.kotlin kotlin-stdlib:latest`
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
jpm_dependencies
*.log
.env
`
}

func GetJavaTestTemplate(packaging string, className string) string {
	return `import org.junit.*;
import ` + packaging + `.` + className + `;
public class Test` + className + ` {
	@Test
	public void test() {
		` + className + ` app = new ` + className + `();
		Assert.assertEquals("Hello, World",app.run());
	}
}`
}

func GetKotlinTestTemplate(packaging string, className string) string {
	return `import org.junit.*
import kotlin.test.*
import ` + packaging + `.` + className + `;
class Test` + className + ` {
	@Test
	fun test() {
		` + className + ` app = ` + className + `();
		assertEquals("Hello, World",app.run());
	}
}`
}

func PrintArt() {
	println(`   ____________  ___
  |_  | ___ \  \/  |  version: 0.0.1
    | | |_/ / .  . |  The simpler
    | |  __/| |\/| |  package manager
/\__/ / |   | |  | |  for your
\____/\_|   \_|  |_/  JVM project`)
}
