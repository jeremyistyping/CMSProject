param(
  [string]$SwaggerJsonPath = "./docs/swagger.json",
  [string]$SwaggerYamlPath = "./docs/swagger.yaml"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Backup-File {
  param([string]$Path)
  if (-not (Test-Path $Path)) { return $null }
  $ts = Get-Date -Format 'yyyyMMdd_HHmmss'
  $backup = [System.IO.Path]::ChangeExtension($Path, ".backup_$ts" + [System.IO.Path]::GetExtension($Path))
  Copy-Item -LiteralPath $Path -Destination $backup -Force
  Write-Host "Backup created: $backup"
  return $backup
}

# Define the basic endpoints to keep (exact matches). Include both with and without /api/v1 prefixes when applicable.
$basicPaths = @(
  # health
  '/api/v1/health',
  '/health',
  # auth
  '/api/v1/auth/login','/auth/login',
  '/api/v1/auth/register','/auth/register',
  '/api/v1/auth/refresh','/auth/refresh',
  '/api/v1/auth/validate-token','/auth/validate-token',
  # profile
  '/api/v1/profile','/profile',
  # cash-bank (read-only basics)
  '/api/v1/cash-bank/accounts',
  '/api/v1/cash-bank/accounts/{id}',
  # journals (read-only basics)
  '/api/v1/journals',
  '/api/v1/journals/{id}'
)

function Load-Json {
  param([string]$Path)
  if (-not (Test-Path $Path)) { throw "Swagger JSON not found: $Path" }
  $raw = Get-Content -LiteralPath $Path -Raw -Encoding UTF8
  return $raw | ConvertFrom-Json
}

function Save-Json {
  param([object]$Obj,[string]$Path)
  ($Obj | ConvertTo-Json -Depth 100) | Out-File -FilePath $Path -Encoding UTF8
}

# Prune swagger.json
Write-Host "Loading Swagger JSON from $SwaggerJsonPath"
$swagger = Load-Json -Path $SwaggerJsonPath
if ($null -eq $swagger.paths) { throw "Invalid swagger.json: no paths" }

$allPaths = @($swagger.paths.PSObject.Properties | ForEach-Object { $_.Name })
$toRemove = @()
foreach ($p in $allPaths) {
  if (-not ($basicPaths -contains $p)) {
    $toRemove += $p
  }
}

Backup-File -Path $SwaggerJsonPath | Out-Null
foreach ($p in $toRemove) {
  $swagger.paths.PSObject.Properties.Remove($p) | Out-Null
}
# Optionally adjust info to note basic-only doc
if ($null -eq $swagger.info) { $swagger | Add-Member -MemberType NoteProperty -Name info -Value (@{}) }
$swagger.info.description = "Basic API documentation (auth, health, profile, read-only essentials). Full endpoints are hidden for now."
Save-Json -Obj $swagger -Path $SwaggerJsonPath
Write-Host "swagger.json pruned to basic endpoints."

# Optionally prune swagger.yaml if yq is available
if (Test-Path $SwaggerYamlPath) {
  $yq = (Get-Command yq -ErrorAction SilentlyContinue)
  if ($yq) {
    Write-Host "Pruning swagger.yaml via yq..."
    Backup-File -Path $SwaggerYamlPath | Out-Null
    # Convert to JSON, edit, convert back
    $jsonTemp = New-TemporaryFile
    (& $yq eval -o=json $SwaggerYamlPath) | Out-File -FilePath $jsonTemp -Encoding UTF8
    $yamlObj = (Get-Content -LiteralPath $jsonTemp -Raw -Encoding UTF8) | ConvertFrom-Json
    if ($yamlObj.paths) {
      $yamlPaths = @($yamlObj.paths.PSObject.Properties | ForEach-Object { $_.Name })
      foreach ($p in $yamlPaths) {
        if (-not ($basicPaths -contains $p)) {
          $yamlObj.paths.PSObject.Properties.Remove($p) | Out-Null
        }
      }
      # update description
      if ($null -eq $yamlObj.info) { $yamlObj | Add-Member -MemberType NoteProperty -Name info -Value (@{}) }
      $yamlObj.info.description = "Basic API documentation (auth, health, profile, read-only essentials)."
      # Save
      ($yamlObj | ConvertTo-Json -Depth 100) | Out-File -FilePath $jsonTemp -Encoding UTF8
      & $yq eval -P $jsonTemp | Out-File -FilePath $SwaggerYamlPath -Encoding UTF8
    }
    Remove-Item $jsonTemp -Force
    Write-Host "swagger.yaml pruned to basic endpoints."
  } else {
    Write-Host "yq not found; skipped swagger.yaml pruning."
  }
}

Write-Host "✅ Done keeping basic API endpoints only."