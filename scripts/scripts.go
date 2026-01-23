package scripts

import (
	"fmt"
	COM "jpm/common"
	"os"
	"slices"
	"strings"
)

var scripts map[string]string

func Scripts(scriptName string) {
	scripts = COM.ParseScripts()
	scriptCmd, found := scripts[scriptName]
	if len(os.Args) > 1 {

	}
	if !found {
		fmt.Printf("Script '%s' not found in package.yml\n", scriptName)
		os.Exit(1)
	}
	cmd := "export PATH=\"$PATH:$(pwd)/jpm_dependencies/execs\"\n" + argsReplacer(scriptCmd)
	if err := COM.RunScript(cmd, true); err != nil {
		fmt.Printf("Error running script '%s': %v \n", scriptName, err)
		os.Exit(1)
	}
}

func argsReplacer(scriptCmd string) string {
	if len(os.Args) > 1 {

		argsSubs := map[string]string{}
		// Replace ...args@ with all args joined as a string
		val := strings.TrimSpace(strings.Join(os.Args[2:], " "))
		if !COM.IsWindows() && !strings.Contains(val, " ") {
			argsSubs["...args@"] = strings.TrimSpace(strings.Join(os.Args[2:], " "))
		} else {
			argsStr := strings.TrimSpace(strings.Join(os.Args[2:], "' '"))
			argsSubs["...args@"] = "'" + argsStr + "'"
		}
		// Replace ...args# with the number of args
		argsSubs["args@[#]"] = fmt.Sprint(len(os.Args) - 2)
		if !COM.IsWindows() && !strings.Contains(argsSubs["args@[#]"], " ") {
			argsSubs["args@[last]"] = os.Args[len(os.Args)-1]
		} else {
			argsSubs["args@[last]"] = "'" + os.Args[len(os.Args)-1] + "'"
		}

		// Replace ...args[i] with the ith arg
		for i, v := range os.Args[2:] {
			if !COM.IsWindows() && !strings.Contains(v, " ") {
				argsSubs["args@["+fmt.Sprint(i)+"]"] = v
			} else {
				argsSubs["args@["+fmt.Sprint(i)+"]"] = "'" + v + "'"
			}
		}

		// Replace ...args[-flag] with true if flag is present
		for _, v := range os.Args[2:] {
			if strings.HasPrefix(v, "-") && !strings.HasPrefix(v, "--") {
				argsSubs["args@["+v+"]"] = "true"
			}
		}
		// Replace ...args[--flag] with value if flag is present
		for i, v := range os.Args[2:] {
			if strings.HasPrefix(v, "--") && !strings.Contains(v, "=") {
				if !COM.IsWindows() && !strings.Contains(v, " ") {
					argsSubs["args@["+v+"]"] = os.Args[2+i+1]
				} else {
					argsSubs["args@["+v+"]"] = "'" + os.Args[2+i+1] + "'"
				}
			}
		}

		// Replace ...args[--flag=value] with value if flag is present
		for _, v := range os.Args[2:] {
			if strings.HasPrefix(v, "--") && strings.Contains(v, "=") {
				if !COM.IsWindows() && !strings.Contains(v, " ") {
					argsSubs["args@["+strings.SplitN(v, "=", 2)[0]+"]"] = strings.SplitN(v, "=", 2)[1]
				} else {
					argsSubs["args@["+strings.SplitN(v, "=", 2)[0]+"]"] = "'" + strings.SplitN(v, "=", 2)[1] + "'"
				}
			}
		}
		// Perform replacements
		for k, v := range argsSubs {
			scriptCmd = strings.ReplaceAll(scriptCmd, k, v)
		}
		// cleanups
		scriptCmdS := strings.Split(scriptCmd, "args@[")
		if len(scriptCmdS) == 1 {
			return scriptCmd
		}
		for i, v := range scriptCmdS {
			if i == 0 {
				continue
			}
			scriptCmdS[i] = trimPrefixUntilClosingBracket(v)
		}
		scriptCmd = strings.Join(scriptCmdS, "")
		if strings.Contains(scriptCmd, "args@") {
			fmt.Println("JPM Error: some args@ placeholders were not replaced properly.")
			os.Exit(1)
		}
		scriptCmd = strings.ReplaceAll(scriptCmd, "args\\@", "args@")
		return scriptCmd
	}
	return scriptCmd
}

func ExecOverride(sc string) {
	if len(scripts) == 0 {
		scripts = COM.ParseScripts()
	}
	scriptsS := []string{}
	for k := range scripts {
		scriptsS = append(scriptsS, k)
	}
	if strings.HasSuffix(os.Args[1], "!") {
		if COM.Verbose {
			println("\033[33mOmitting Override: "+"'"+"jpm", sc+"'", "\033[0m")
		}
		return
	}
	if os.Getenv("JPM_OVERRIDE") != sc {
		// Join all the args from os.Args except os.Args[0] into a string
		argsStr := ""
		if len(os.Args) > 1 {
			argsStr = strings.TrimSpace(strings.Join(os.Args[2:], " "))
		}
		if slices.Contains(scriptsS, sc+"@") {
			if COM.Verbose {
				println("\033[33mOverriding: "+"'"+"jpm", sc+"'", "for "+"'"+"jpm", sc+"@"+"'", "\033[0m")
			}
			cmd := "export JPM_OVERRIDE=" + sc + "\nexport PATH=\"$PATH:$(pwd)/jpm_dependencies/execs\"\n" + strings.ReplaceAll(scripts[sc+"@"], "...args@", argsStr)
			if err := COM.RunScript(cmd, true); err != nil {
				fmt.Printf("Error running script '%s': %v \n", sc, err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}
}

func trimPrefixUntilClosingBracket(s string) string {
	idx := strings.Index(s, "]")
	if idx == -1 {
		return s
	}
	return s[idx+1:]
}
