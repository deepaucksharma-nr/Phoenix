@echo off
REM Phoenix E2E Dependencies and Contracts Validation Script (Windows Batch)

echo === Phoenix E2E Dependencies and Contracts Validation ===
echo.

REM Check Go version
echo Checking Go version...
go version
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Go is not installed
    exit /b 1
)

echo.
echo Validating Go modules...

REM Validate pkg module
echo Checking pkg module...
cd pkg
go mod verify > nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [OK] pkg module dependencies verified
) else (
    echo [ERROR] pkg module dependencies verification failed
    cd ..
    exit /b 1
)
cd ..

echo.
echo Checking contracts...

REM Check OpenAPI contracts
if exist "pkg\contracts\openapi\control-api.yaml" (
    echo [OK] Found OpenAPI contract: control-api.yaml
) else (
    echo [ERROR] OpenAPI contract not found
    exit /b 1
)

REM Check Proto contracts
if exist "pkg\contracts\proto\v1\common.proto" (
    echo [OK] Found Proto contract: common.proto
) else (
    echo [ERROR] Proto contract not found: common.proto
)

if exist "pkg\contracts\proto\v1\controller.proto" (
    echo [OK] Found Proto contract: controller.proto
) else (
    echo [ERROR] Proto contract not found: controller.proto
)

if exist "pkg\contracts\proto\v1\experiment.proto" (
    echo [OK] Found Proto contract: experiment.proto
) else (
    echo [ERROR] Proto contract not found: experiment.proto
)

if exist "pkg\contracts\proto\v1\generator.proto" (
    echo [OK] Found Proto contract: generator.proto
) else (
    echo [ERROR] Proto contract not found: generator.proto
)

echo.
echo Checking E2E test files...

REM Check E2E tests
if exist "tests\e2e\simple_e2e_test.go" (
    echo [OK] Found E2E test: simple_e2e_test.go
) else (
    echo [ERROR] E2E test not found: simple_e2e_test.go
)

if exist "tests\e2e\experiment_workflow_test.go" (
    echo [OK] Found E2E test: experiment_workflow_test.go
) else (
    echo [ERROR] E2E test not found: experiment_workflow_test.go
)

echo.
echo Testing E2E compilation...
cd tests\e2e
go test -tags e2e -c > nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [OK] E2E tests compile successfully
    if exist e2e.test.exe del e2e.test.exe
) else (
    echo [ERROR] E2E tests compilation failed
    echo Run 'go test -tags e2e -c' in tests\e2e for details
)
cd ..\..

echo.
echo === E2E Dependencies and Contracts Validation Complete ===
echo To run E2E tests, use: make test-e2e
