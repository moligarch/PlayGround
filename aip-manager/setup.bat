@echo off
setlocal enabledelayedexpansion

REM ============================================================================
REM == AIP Manager Project Script (v2)
REM ============================================================================

REM --- Interactive Menu Mode ---
:menu
cls
echo.
echo ============================================================
echo  AIP Manager - Project Control
echo ============================================================
echo.
echo  Select an option:
echo.
echo    [1] Run All Go Tests
echo    [2] Build Manager.exe
echo.
echo    [0] Exit
echo.
set /p "choice=Enter your choice: "

if "%choice%"=="1" goto test
if "%choice%"=="2" goto package
if "%choice%"=="0" EXIT /B 0
echo Invalid choice.
goto menu

REM --- Test Logic ---
:test
echo.
echo --- [1] Running All Go Tests ---
go test -v ./...
goto end

REM --- Package Logic ---
:package
echo.
echo --- [2] Building Manager.exe (%AIP_MANAGER_VERSION%) ---
go build -ldflags "-s -w" -o "./build/Manager.exe" .
if errorlevel 1 (
    echo.
    echo FAILED to build executable.
) else (
    echo.
    echo Successfully created "Manager.exe"
)
goto end

:end
echo.
endlocal
