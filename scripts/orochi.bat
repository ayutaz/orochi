@echo off
:: Orochi BitTorrent Client Launcher for Windows
:: This script ensures the application runs properly on Windows

:: Get the directory where this script is located
set SCRIPT_DIR=%~dp0
set OROCHI_EXE=%SCRIPT_DIR%..\orochi.exe

:: Check if orochi.exe exists
if not exist "%OROCHI_EXE%" (
    echo Error: orochi.exe not found in %SCRIPT_DIR%..
    echo Please ensure the binary is in the correct location.
    pause
    exit /b 1
)

:: Run Orochi with any passed arguments
echo Starting Orochi BitTorrent Client...
"%OROCHI_EXE%" %*

:: Keep console open if there was an error
if %ERRORLEVEL% NEQ 0 (
    echo.
    echo Orochi exited with error code: %ERRORLEVEL%
    pause
)