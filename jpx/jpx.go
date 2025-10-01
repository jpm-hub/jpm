package jpx

func SHTemplate() string {
	return ` #!/bin/bash
export PATH="$PATH:$(pwd)/jpm_dependencies/execs"
if [ $# -eq 0 ]; then
	jpm
else
	"$@"
fi`
}

func CMDTemplate() string {
	return `@echo off
setlocal
set "PATH=%PATH%;%CD%\jpm_dependencies\execs"
if "%*"=="" (
	jpm
) else (
	%*
)
endlocal`
}
