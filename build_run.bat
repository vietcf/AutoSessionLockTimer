@echo off
setlocal enabledelayedexpansion

echo [1/3] Checking Go...
go version >nul 2>&1
if errorlevel 1 (
  echo Go is not installed or not in PATH.
  echo Install Go from https://go.dev/dl/ and try again.
  exit /b 1
)

echo [2/3] Building AutoLock.exe (windowsgui)...
go build -ldflags="-H=windowsgui" -o AutoLock.exe
if errorlevel 1 (
  echo Build failed.
  exit /b 1
)

echo [3/3] Running AutoLock.exe...
tasklist /FI "IMAGENAME eq AutoLock.exe" | find /I "AutoLock.exe" >nul
if not errorlevel 1 (
  echo AutoLock.exe is already running. Skip starting a new instance.
  exit /b 0
)
start "" "%cd%\AutoLock.exe"
exit /b 0
