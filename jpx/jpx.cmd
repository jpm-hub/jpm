@echo off
setlocal
set "PATH=%PATH%;%CD%\jpm_dependencies\execs"
if "%*"=="" (
	jpm
) else (
	%*
)
endlocal
