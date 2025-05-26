# Phoenix E2E Dependencies and Contracts Validation Script (PowerShell)
# This script validates all dependencies and contracts required for e2e testing

$ErrorActionPreference = "Stop"

# Colors for output
function Write-Success { 
    param([string]$Message)
    Write-Host $Message -ForegroundColor Green 
}

function Write-Error { 
    param([string]$Message)
    Write-Host $Message -ForegroundColor Red 
}

function Write-Warning { 
    param([string]$Message)
    Write-Host $Message -ForegroundColor Yellow 
}

function Write-Info { 
    param([string]$Message)
    Write-Host $Message -ForegroundColor Cyan 
}

Write-Info "=== Phoenix E2E Dependencies and Contracts Validation ==="
Write-Host ""

# Function to check if a command exists
function Test-Command {
    param($Command)
    try {
        if (Get-Command $Command -ErrorAction Stop) {
            return $true
        }
    }
    catch {
        return $false
    }
}

# Function to check Go version
function Test-GoVersion {
    if (Test-Command "go") {
        $goVersion = go version
        $versionMatch = [regex]::Match($goVersion, 'go(\d+\.\d+\.\d+)')
        if ($versionMatch.Success) {
            $version = $versionMatch.Groups[1].Value
            $requiredVersion = "1.24.0"
            Write-Success "✓ Go version $version found"
            # Simple version comparison
            try {
                if ([version]$version -ge [version]$requiredVersion) {
                    Write-Success "✓ Go version meets requirement (>= $requiredVersion)"
                }
                else {
                    Write-Error "✗ Go version does not meet requirement (>= $requiredVersion)"
                    exit 1
                }
            }
            catch {
                Write-Warning "Could not compare versions, assuming compatible"
            }
        }
    }
    else {
        Write-Error "✗ Go is not installed"
        exit 1
    }
}

# Function to validate Go modules
function Test-GoModules {
    Write-Warning "`nValidating Go modules..."
    
    # Check main pkg module
    if (Test-Path "pkg\go.mod") {
        Write-Success "✓ Found pkg\go.mod"
        Push-Location pkg
        try {
            $result = go mod verify 2>&1
            if ($LASTEXITCODE -eq 0) {
                Write-Success "✓ pkg module dependencies verified"
            }
            else {
                Write-Error "✗ pkg module dependencies verification failed"
                exit 1
            }
        }
        finally {
            Pop-Location
        }
    }
    else {
        Write-Error "✗ pkg\go.mod not found"
        exit 1
    }
    
    # Check all project modules
    Get-ChildItem -Path "projects" -Directory | ForEach-Object {
        $projectPath = $_.FullName
        $projectName = $_.Name
        if (Test-Path "$projectPath\go.mod") {
            Write-Success "✓ Found $projectName\go.mod"
            Push-Location $projectPath
            try {
                $result = go mod verify 2>&1
                if ($LASTEXITCODE -eq 0) {
                    Write-Success "✓ $projectName module dependencies verified"
                }
                else {
                    Write-Error "✗ $projectName module dependencies verification failed"
                    exit 1
                }
            }
            finally {
                Pop-Location
            }
        }
    }
}

# Function to check contracts
function Test-Contracts {
    Write-Warning "`nChecking contracts..."
    
    # Check OpenAPI contracts
    if (Test-Path "pkg\contracts\openapi\control-api.yaml") {
        Write-Success "✓ Found OpenAPI contract: control-api.yaml"
    }
    else {
        Write-Error "✗ OpenAPI contract not found"
        exit 1
    }
    
    # Check Proto contracts
    $protoFiles = @(
        "pkg\contracts\proto\v1\common.proto",
        "pkg\contracts\proto\v1\controller.proto",
        "pkg\contracts\proto\v1\experiment.proto",
        "pkg\contracts\proto\v1\generator.proto"
    )
    
    foreach ($proto in $protoFiles) {
        if (Test-Path $proto) {
            Write-Success "✓ Found Proto contract: $(Split-Path $proto -Leaf)"
        }
        else {
            Write-Error "✗ Proto contract not found: $proto"
            exit 1
        }
    }
}

# Function to check E2E test files
function Test-E2ETests {
    Write-Warning "`nChecking E2E test files..."
    
    $e2eTests = @(
        "tests\e2e\simple_e2e_test.go",
        "tests\e2e\experiment_workflow_test.go"
    )
    
    foreach ($test in $e2eTests) {
        if (Test-Path $test) {
            Write-Success "✓ Found E2E test: $(Split-Path $test -Leaf)"
            # Check for e2e build tag
            $content = Get-Content $test -Raw
            if ($content -match "//go:build e2e" -or $content -match "// \+build e2e") {
                Write-Success "  ✓ Has e2e build tag"
            }
            else {
                Write-Warning "  ⚠ Missing e2e build tag"
            }
        }
        else {
            Write-Error "✗ E2E test not found: $test"
            exit 1
        }
    }
}

# Function to check required services for E2E
function Test-RequiredServices {
    Write-Warning "`nChecking required services configuration..."
    
    # Check docker-compose.yml
    if (Test-Path "docker-compose.yml") {
        Write-Success "✓ Found docker-compose.yml"
        $dockerCompose = Get-Content "docker-compose.yml" -Raw
        if ($dockerCompose -match "postgres") {
            Write-Success "  ✓ PostgreSQL service configured"
        }
        else {
            Write-Warning "  ⚠ PostgreSQL service not found in docker-compose"
        }
    }
    else {
        Write-Warning "⚠ docker-compose.yml not found"
    }
    
    # Check for service configurations
    $services = @(
        "platform-api",
        "controller",
        "pipeline-operator"
    )
    
    foreach ($service in $services) {
        if (Test-Path "projects\$service") {
            Write-Success "✓ Found service: $service"
        }
        else {
            Write-Warning "⚠ Service directory not found: $service"
        }
    }
}

# Function to validate E2E test dependencies
function Test-E2EDependencies {
    Write-Warning "`nValidating E2E test dependencies..."
    
    # Required Go packages for E2E tests
    $requiredPackages = @(
        "github.com/stretchr/testify",
        "github.com/google/uuid",
        "google.golang.org/grpc",
        "k8s.io/client-go"
    )
    
    # Get all go.mod files
    $goModFiles = @()
    if (Test-Path "pkg\go.mod") {
        $goModFiles += Get-Item "pkg\go.mod"
    }
    $projectMods = Get-ChildItem -Path "projects" -Filter "go.mod" -Recurse -ErrorAction SilentlyContinue
    if ($projectMods) {
        $goModFiles += $projectMods
    }
    
    $foundPackages = 0
    foreach ($pkg in $requiredPackages) {
        $found = $false
        foreach ($modFile in $goModFiles) {
            if ($modFile -and (Get-Content $modFile.FullName -Raw) -match [regex]::Escape($pkg)) {
                $found = $true
                break
            }
        }
        
        if ($found) {
            Write-Success "✓ Found dependency: $pkg"
            $foundPackages++
        }
        else {
            Write-Warning "⚠ Dependency not found in go.mod files: $pkg"
        }
    }
    
    if ($foundPackages -eq $requiredPackages.Count) {
        Write-Success "✓ All E2E test dependencies found"
    }
    else {
        Write-Warning "⚠ Some E2E test dependencies might be missing"
    }
}
