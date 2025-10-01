package watch

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	COM "jpm/common"
	COMPILE "jpm/compile"

	"github.com/fsnotify/fsnotify"
)

var command string = "jpm compile"
var procs []*os.Process = []*os.Process{}

func Watch(fromRun bool) {
	COM.FindPackageYML(true)
	// Create a new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Error creating watcher:", err)
		return
	}
	defer watcher.Close()

	// Parse the filter pattern
	patterns := parseFilterPattern(figureOutFilter(fromRun))
	fmt.Println("\033[33mWatching file changes " + figureOutFilter(fromRun) + "\033[0m")
	if len(patterns) == 0 {
		fmt.Println("Invalid filter pattern")
		return
	}
	if fromRun {
		println("\n\n\033[33m Warning: 'jpm run -hot' is in alpha, hot reloading may not work as expected\033[0m\n")
	}
	// Start watching for changes
	go func(fromRunWatch bool) {
		for {
			event, ok := <-watcher.Events
			if !ok {
				break
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				if matchesPattern(event.Name, patterns) {
					refresh(fromRunWatch)
					time.Sleep(time.Second)
				}
			}
		}
	}(fromRun)

	// Add directories to watch based on the filter pattern
	dirs := getDirectoriesToWatch(patterns)
	for _, dir := range dirs {
		err = watcher.Add(dir)
		if err != nil {
			fmt.Printf("Error watching directory %s: %v\n", dir, err)
		}
	}

	// Wait for interrupt signal (Ctrl+C) and run killpids() before exiting
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
	killpids()
}

func killpids() {
	for _, proc := range procs {
		if proc == nil {
			continue
		}
		killProcessGroup(proc)
	}
}

func figureOutFilter(fromRun bool) string {
	src := COM.SrcDir()
	switch fromRun {
	case false:
		if len(os.Args) == 2 {
			return "(" + src + "/**)"
		} else if len(os.Args) == 3 && os.Args[2] != "_" {
			return os.Args[2]
		} else if len(os.Args) == 4 {
			if os.Args[3] != "_" {
				command = os.Args[3]
			}
			if os.Args[2] != "_" {
				return os.Args[2]
			}
		}
	case true:
		command = ""
		if len(os.Args) == 3 {
			return "(" + src + "/**)"
		} else if len(os.Args) == 4 && os.Args[3] != "_" {
			return os.Args[3]
		} else if len(os.Args) > 4 {
			if os.Args[4] != "_" {
				command = os.Args[4]
			}
			if os.Args[3] != "_" {
				return os.Args[3]
			}
		}
	}
	return "(" + src + "/**)"
}

func parseFilterPattern(filter string) []string {
	// Remove parentheses if present
	filter = strings.Trim(filter, "()")

	// Split by pipe for multiple patterns
	patterns := strings.Split(filter, "|")

	// Process each pattern
	var processedPatterns []string
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if strings.HasPrefix(pattern, "!") {
			// Handle exclusion patterns
			pattern = strings.TrimPrefix(pattern, "!")
			processedPatterns = append(processedPatterns, "!"+pattern)
		} else {
			processedPatterns = append(processedPatterns, pattern)
		}
	}

	return processedPatterns
}

func matchesPattern(filePath string, patterns []string) bool {
	// Convert file path to use forward slashes
	filePath = filepath.ToSlash(filePath)

	for _, pattern := range patterns {
		isExclusion := strings.HasPrefix(pattern, "!")
		if isExclusion {
			pattern = strings.TrimPrefix(pattern, "!")
		}

		// Convert glob pattern to regex
		regexPattern := convertGlobToRegex(pattern)
		matched, err := regexp.MatchString(regexPattern, filePath)
		if err != nil {
			continue
		}

		if isExclusion {
			if matched {
				return false
			}
		} else {
			if matched {
				return true
			}
		}
	}

	return false
}

func convertGlobToRegex(pattern string) string {
	// Replace glob patterns with regex patterns
	pattern = regexp.QuoteMeta(pattern)
	pattern = strings.ReplaceAll(pattern, "\\*\\*", ".*") // ** matches any number of directories
	pattern = strings.ReplaceAll(pattern, "\\*", "[^/]*") // * matches any characters except /
	return "^" + pattern + "$"
}

func getDirectoriesToWatch(patterns []string) []string {
	var dirs []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		if strings.HasPrefix(pattern, "!") {
			continue
		}

		// Extract the base directory from the pattern
		parts := strings.Split(pattern, "/")
		if len(parts) > 0 {
			dir := parts[0]
			if !seen[dir] {
				dirs = append(dirs, dir)
				seen[dir] = true

				// Recursively add all subdirectories
				filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						if !seen[path] {
							dirs = append(dirs, path)
							seen[path] = true
						}
					}
					return nil
				})
			}
		}
	}
	return dirs
}

func refresh(fromRunWatch bool) {
	var proc *os.Process
	if !fromRunWatch {
		killpids()
		fmt.Println("\033[32mRunning: " + command + "\033[0m")
		proc = runScriptWithPID(command)
	} else {
		if command != "" {
			println("\033[32mexecuting : " + command + "\033[0m")
			COM.RunScript(command, true)
		}
		println("\n\033[32mCompiling\033[0m")
		COMPILE.Compile()
	}
	procs = append(procs, proc)
}

func runScriptWithPID(script string) *os.Process {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// On Windows, use cmd.exe to create a process group
		cmd = exec.Command("cmd", "/C", script)
	} else {
		// On Unix, use sh -c with exec to ensure proper process group
		cmd = exec.Command("sh", "-c", script)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Set process group attributes (platform-specific)
	setProcessGroupAttr(cmd)

	err := cmd.Start()
	if err != nil {
		return nil
	}
	return cmd.Process
}
