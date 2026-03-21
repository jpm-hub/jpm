@echo off
setlocal enabledelayedexpansion
set VERSION=1.2.9

:: JPM Windows Setup Script
echo ===============================================
echo      JPM Installation Script (v%VERSION%)
echo ===============================================
echo.

:: Get current user
for /f "tokens=2 delims==" %%a in ('wmic useraccount where name="%USERNAME%" get SID /value 2^>nul') do set USER_SID=%%a 2>&1

:: Set JPM directories
set JPM_HOME=C:\Users\%USERNAME%\.jpm

echo This script will install JPM %VERSION% 
echo.
echo The following actions will be performed:
echo.
echo 1. Create directories:
echo    - %JPM_HOME%
echo.
echo 2. Extract JPM files from the downloaded ZIP from github release
echo    here : https://github.com/jpm-hub/jpm/releases
echo.
echo 3. Remove old JPM files (if they exist):
echo    - %JPM_HOME%\jpm.exe
echo.
echo 4. Add the following directories to your user PATH:
echo    - %JPM_HOME%
echo.
pause

:: Create JPM directories
echo Creating JPM directories...
if not exist "%JPM_HOME%" mkdir "%JPM_HOME%"
echo - Directories created successfully
echo.

:: Detect system architecture
echo Detecting system architecture...
for /f "tokens=2 delims==" %%a in ('wmic os get osarchitecture /value') do set ARCH=%%a
set ARCH=%ARCH:~0,-1%

if "%PROCESSOR_ARCHITECTURE%"=="ARM64" (
    set ARCH_TYPE=arm64
    echo - Detected ARM64 architecture
) else if "%PROCESSOR_ARCHITECTURE%"=="AMD64" (
    set ARCH_TYPE=amd64
    echo - Detected AMD64 architecture
) else if "%ARCH%"=="64-bit" (
    set ARCH_TYPE=amd64
    echo - Detected 64-bit x86 architecture
) else (
    set ARCH_TYPE=amd64
    echo ! Could not detect architecture, defaulting to AMD64
)

:: Download JPM ZIP file
echo Downloading JPM from GitHub releases...
set ZIP_FILE=jpm-windows-%ARCH_TYPE%.zip
set DOWNLOAD_URL=https://github.com/jpm-hub/jpm/releases/download/v%VERSION%/jpm-windows-%ARCH_TYPE%.zip
:: Remove existing ZIP file if it exists
if exist "%ZIP_FILE%" del "%ZIP_FILE%"

:: Download with curl
curl --location -o %JPM_HOME%\%ZIP_FILE% %DOWNLOAD_URL%
if %errorLevel% neq 0 (
    echo ERROR: Failed to download JPM ZIP file
    echo Please check your internet connection and try again.
    pause
    exit /b 1
)
echo - Downloaded %ZIP_FILE% successfully

:: Remove old JPM files
echo.
echo Removing old JPM files (if they exist)...
if exist "%JPM_HOME%\jpm.exe" (
    del "%JPM_HOME%\jpm.exe"
    echo - Removed old jpm.exe
)
if exist "%JPM_HOME%\jpx.cmd" (
    del "%JPM_HOME%\jpx.cmd"
    echo - Removed old jpx.cmd
)
if exist "%JPM_HOME%\jpm" (
    del "%JPM_HOME%\jpm"
    echo - Removed old jpm
)
echo.

:: Extract ZIP file
echo Extracting %ZIP_FILE%...
cd %JPM_HOME%
C:\Windows\System32\tar.exe -xf %ZIP_FILE%
if %errorLevel% neq 0 (
    echo ERROR: Failed to extract ZIP file
    pause
    exit /b 1
)
del %ZIP_FILE%
echo - Files extracted successfully
echo.
echo.
echo ===============================================
echo           Installation Complete!
echo ===============================================
echo.
echo Please add %JPM_HOME% to your PATH environment variable.
echo run : win + r > SystemPropertiesAdvanced.exe 
echo              > Environment Variables 
echo              > User variables > Path > Edit 
echo              > New > %JPM_HOME%
echo.
echo IMPORTANT: You may need to restart your Command Prompt or
echo            PowerShell for PATH changes to take effect.
echo.
echo To verify installation, open a new Command Prompt and run:
echo   jpm 
echo   jpm -h
echo.
pause
