@echo off
setlocal
set find_package_yml=0
set max_traverse=20
set traverse_count=0
set curdir=%CD%

set arg_count=0
for %%x in (%*) do set /a arg_count+=1
set "PATH=%CD%\jpm_dependencies\execs;%PATH%"
:find_package_yml_loop
if exist "%curdir%\package.yml" (
  set find_package_yml=1
  cd /d "%curdir%"
  goto :after_find_package_yml
)
set /a traverse_count+=1
if %traverse_count% GEQ %max_traverse% (
  goto :after_find_package_yml
)
for %%A in ("%curdir%\..") do set curdir=%%~fA
if "%curdir%" == "%%~dpcd%%" (
  goto :after_find_package_yml
)
goto :find_package_yml_loop

:after_find_package_yml
if %find_package_yml%==0 (
  echo Error: package.yml not found
  exit /b
)
if %arg_count%==0 (
  echo Usage: jpx <command> [args...]
  echo available commands :
  "C:\Program Files\Git\bin\bash.exe" -c "for file in ""$(pwd)/jpm_dependencies/execs/""*; do if [ -x ""$file"" ] && [ -f ""$file"" ]; then echo "" > $(basename ""$file"")""; fi; done;"
  ) else (
  if "%1"=="-d" (
    "C:\Program Files\Git\bin\bash.exe" -c "cmd=%*; cmd=${cmd:3}; ./jpm_dependencies/execs/${cmd} > /dev/null 2>&1 &"
    echo Started in detached mode: %*
  ) else (
    %*
  )
)
endlocal