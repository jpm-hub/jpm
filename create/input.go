package create

import (
	"fmt"
	"maps"
	"os"
	"strings"
)

func inputText(s string) string {
	if len(s) >= len("<input-text>") && s[len(s)-len("<input-text>"):] == "<input-text>" {
		s = s[:len(s)-len("<input-text>")]
	}
	var input string
	print(s)
	// Read input until newline
	fmt.Scanln(&input)
	return input
}

func inputNumber(s string) string {
	if len(s) >= len("<input-number>") && s[len(s)-len("<input-number>"):] == "<input-number>" {
		s = s[:len(s)-len("<input-number>")]
	}
	var input string
	var number int
	for {
		print(s)
		fmt.Scanln(&input)
		_, err := fmt.Sscanf(input, "%d", &number)
		if err == nil {
			break
		}
		fmt.Println("\033[31mPlease enter a valid number.\033[0m")
	}
	return fmt.Sprintf("%d", number)
}

func inputChoice(key string, s string) map[string]string {
	if len(s) >= len("<input-choice>") && s[len(s)-len("<input-choice>"):] == "<input-choice>" {
		s = s[:len(s)-len("<input-choice>")]
	}
	var input string
	choice := map[string]any{}
	println(s, "\n")
	for v := range maps.Keys(templateYml) {
		prefix := "<choice>:" + key + ":"
		if strings.HasPrefix(v, prefix) {
			suffix := strings.TrimPrefix(v, prefix)

			if val, ok := templateYml[v].(map[any]any); ok {
				println("   ", suffix, "-", val["title"].(string))
				choice[suffix] = val
			} else {
				fmt.Println(v, "should be a map")
				os.Exit(1)
			}
		}
	}
	// Read input until newline
	print("   choice: ")
	fmt.Scanln(&input)
	for {
		if val, ok := choice[input].(map[any]any); ok {
			value := map[string]string{}
			for k, v := range val {
				value[k.(string)] = v.(string)
			}
			return value
		} else {
			fmt.Println("\033[31mPlease enter a valid choice.\033[0m")
			fmt.Scanln(&input)
		}
	}

}
