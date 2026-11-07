param(
    [ValidateSet("basic", "integration", "reference", "smoke", "performance", "all")]
    [string]$Suite = "basic",
    [switch]$List,
    [string]$Match = "",
    [string]$Timeout = ""
)
$ErrorActionPreference = "Stop"
$testArgs = @((Join-Path $PSScriptRoot "tools/run_tests.py"), "--suite", $Suite)
if ($List) { $testArgs += "--list" }
if ($Match) { $testArgs += @("--match", $Match) }
if ($Timeout) { $testArgs += @("--timeout", $Timeout) }
& python @testArgs
exit $LASTEXITCODE
