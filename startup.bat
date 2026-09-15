@echo off
setlocal

echo ========================================
echo       NetPilot Development Startup
echo ========================================
echo.

echo [1/7] Formatting Go code...
go fmt ./...
if errorlevel 1 (
    echo ERROR: go fmt failed.
    exit /b 1
)

echo.
echo [2/7] Running golines...
golines -w -m 120 .
if errorlevel 1 (
    echo ERROR: golines failed.
    exit /b 1
)

echo.
echo [3/7] Running tests...
go test ./...
if errorlevel 1 (
    echo ERROR: go test failed.
    exit /b 1
)

echo.
echo [4/7] Building NetPilot API...
if not exist bin mkdir bin

go build -o bin\netpilot-api.exe .\cmd\api
if errorlevel 1 (
    echo ERROR: NetPilot API build failed.
    exit /b 1
)

echo.
echo [5/7] Setting bootstrap admin environment variables...

set "NETPILOT_BOOTSTRAP_ADMIN_NAME=NetPilot Admin"
set "NETPILOT_BOOTSTRAP_ADMIN_EMAIL=admin@netpilot.local"
set "NETPILOT_BOOTSTRAP_ADMIN_PASSWORD=AdminNetPilot@1234!"

echo.
echo [6/7] Running admin bootstrap...

go run .\cmd\admin-bootstrap

if errorlevel 1 (
    echo.
    echo NOTE: Admin bootstrap did not create an admin.
    echo This is expected if an administrator already exists.
)

echo.
echo Clearing bootstrap environment variables...

set "NETPILOT_BOOTSTRAP_ADMIN_NAME="
set "NETPILOT_BOOTSTRAP_ADMIN_EMAIL="
set "NETPILOT_BOOTSTRAP_ADMIN_PASSWORD="

echo.
echo [7/7] Starting NetPilot API...
echo ========================================
echo.

.\bin\netpilot-api.exe