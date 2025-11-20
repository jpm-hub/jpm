@echo off
setlocal
set find_package_yml=0
set max_traverse=10
set traverse_count=0
set curdir=%CD%

set arg_count=0
for %%x in (%*) do set /a arg_count+=1
if %arg_count%==0 (
  "C:\Program Files\Git\bin\bash.exe" -c "echo Usage: jpx ^<command^> [args...]; echo available commands :; for file in ""$(pwd)/jpm_dependencies/execs/""*; do if [ -x ""$file"" ] && [ -f ""$file"" ]; then echo $(basename ""$file""); fi; done"
  ) else (
  set "PATH=%PATH%;%CD%\jpm_dependencies\execs"
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
    echo "Error: package.yml not found"
    exit /b
  )
  "C:\Program Files\Git\bin\bash.exe" %*
)
endlocal