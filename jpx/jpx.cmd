@echo off
setlocal
set "PATH=%PATH%;%CD%\jpm_dependencies\execs"
set arg_count=0
for %%x in (%*) do set /a arg_count+=1
if %arg_count%==0 (
        jpm
) else (
        "C:\Program Files\Git\bin\bash.exe" %*
)
endlocal
