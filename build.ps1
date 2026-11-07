param(
    [string]$Output = "bin/atlas.exe"
)

$ErrorActionPreference = "Stop"
$goCommand = Get-Command go -ErrorAction SilentlyContinue
if ($goCommand) {
    $goExe = $goCommand.Source
} else {
    $goExe = Join-Path $env:TEMP "atlas-go\go\bin\go.exe"
    if (-not (Test-Path -LiteralPath $goExe)) {
        throw "Go was not found in PATH or $goExe"
    }
}

$buildTime = [DateTime]::UtcNow.ToString("yyyy-MM-ddTHH:mm:ssZ")
if (-not [IO.Path]::IsPathRooted($Output)) {
    $Output = Join-Path $PSScriptRoot $Output
}
$outputDirectory = Split-Path -Parent $Output
New-Item -ItemType Directory -Path $outputDirectory -Force | Out-Null
Push-Location $PSScriptRoot
try {
    & $goExe build -ldflags "-X main.buildTime=$buildTime" -o $Output ./cmd/atlas
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
} finally {
    Pop-Location
}
Write-Host "Built $Output at $buildTime (UTC)"
