param(
    [string[]]$ServerArgs = @(),
    [switch]$Rebuild
)
$ErrorActionPreference = "Stop"
$executable = Join-Path $PSScriptRoot "bin/atlas.exe"
$runDirectory = Join-Path $PSScriptRoot "run"
if ($Rebuild -or -not (Test-Path -LiteralPath $executable)) {
    & (Join-Path $PSScriptRoot "build.ps1")
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}
New-Item -ItemType Directory -Path $runDirectory -Force | Out-Null
$properties = Join-Path $runDirectory "server.properties"
if (-not (Test-Path -LiteralPath $properties)) {
    Copy-Item -LiteralPath (Join-Path $PSScriptRoot "configs/server.example.properties") -Destination $properties
}
Write-Host "Runtime directory: $runDirectory"
Write-Host "Type stop to save and shut down."
Push-Location $runDirectory
try {
    & $executable @ServerArgs
    $serverExit = $LASTEXITCODE
} finally {
    Pop-Location
}
exit $serverExit
