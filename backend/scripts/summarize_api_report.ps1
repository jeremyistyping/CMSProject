param([string]$ReportPath = "./API_TEST_REPORT.json")

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

if (-not (Test-Path $ReportPath)) { throw "Report not found: $ReportPath" }
$r = Get-Content -LiteralPath $ReportPath -Raw | ConvertFrom-Json
$tested = @($r.results | Where-Object { $_.tested -eq $true })
$pass = @($tested | Where-Object { $_.ok -eq $true })
$fail = @($tested | Where-Object { $_.ok -ne $true })

Write-Host ("Tested: {0}" -f $tested.Count)
Write-Host ("Passed: {0}" -f $pass.Count)
Write-Host ("Failed: {0}" -f $fail.Count)
Write-Host "\nFirst 20 failures:"
$fail | Select-Object -First 20 path,method,status,uri,error | Format-Table -AutoSize
