param(
    [ValidateSet("basic", "reference", "smoke", "performance", "all")]
    [string]$Suite = "basic",
    [switch]$List,
    [string]$Match = ""
)
$ErrorActionPreference = "Stop"
$testArgs = @((Join-Path $PSScriptRoot "tools/run_tests.py"), "--suite", $Suite)
if ($List) { $testArgs += "--list" }
if ($Match) { $testArgs += @("--match", $Match) }
& python @testArgs
exit $LASTEXITCODE
