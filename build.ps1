param(
    [string]$Output = "atlas.exe"
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
& $goExe build -ldflags "-X main.buildTime=$buildTime" -o $Output ./cmd/atlas
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
Write-Host "Built $Output at $buildTime (UTC)"
