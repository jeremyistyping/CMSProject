param(
  [string]$BaseUrl = "http://localhost:8080",
  [string]$SwaggerJsonPath = "./docs/swagger.json",
  [string]$SwaggerYamlPath = "./docs/swagger.yaml",
  [string]$AuthEmail = "admin@company.com",
  [string]$AuthPassword = "password123",
  [switch]$Clean = $false,
  [string]$ReportPath = "./API_TEST_REPORT.json"
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

function Get-AccessToken {
  param([string]$BaseUrl,[string]$Email,[string]$Password)
  $uri = "$BaseUrl/api/v1/auth/login"
  $body = @{ email = $Email; password = $Password } | ConvertTo-Json -Depth 5
  try {
    $resp = Invoke-RestMethod -Method POST -Uri $uri -ContentType 'application/json' -Body $body
    if ($resp.access_token) { return $resp.access_token }
    if ($resp.token) { return $resp.token }
    throw "No access_token field in response"
  } catch {
    throw "Login failed: $($_.Exception.Message) at $uri"
  }
}

function New-RandomEmail {
  $ts = [DateTime]::UtcNow.ToString('yyyyMMddHHmmss')
  return "tester+$ts@company.com"
}

function Try-RegisterAndGetToken {
  param([string]$BaseUrl)
  $uri = "$BaseUrl/api/v1/auth/register"
  $email = New-RandomEmail
  $password = 'password123'
  $bodyObj = @{ 
    username = $email.Split('@')[0]
    email = $email
    password = $password
    first_name = 'Tester'
    last_name = 'Auto'
    role = 'employee'
  }
  $body = $bodyObj | ConvertTo-Json -Depth 5
  try {
    $resp = Invoke-RestMethod -Method POST -Uri $uri -ContentType 'application/json' -Body $body
    if ($resp.access_token) { return @{ token=$resp.access_token; email=$email; password=$password } }
    if ($resp.token) { return @{ token=$resp.token; email=$email; password=$password } }
    return $null
  } catch {
    return $null
  }
}

function Load-SwaggerJson {
  param([string]$Path)
  if (-not (Test-Path $Path)) { throw "Swagger JSON not found: $Path" }
  $raw = Get-Content -LiteralPath $Path -Raw -Encoding UTF8
  return $raw | ConvertFrom-Json
}

function Save-SwaggerJson {
  param([object]$Swagger,[string]$Path)
  ($Swagger | ConvertTo-Json -Depth 100) | Out-File -FilePath $Path -Encoding UTF8
}

function Load-SwaggerYaml {
  param([string]$Path)
  if (-not (Test-Path $Path)) { return $null }
  try {
    Add-Type -AssemblyName System.Web
  } catch {}
  # Simple YAML loader by calling yq if available, else skip YAML
  $yq = (Get-Command yq -ErrorAction SilentlyContinue)
  if ($yq) {
    $json = & $yq eval -o=json $Path
    return $json | ConvertFrom-Json
  }
  Write-Host "yq not found; YAML cleaning will be skipped"
  return $null
}

function Save-SwaggerYaml {
  param([object]$Swagger,[string]$Path)
  $yq = (Get-Command yq -ErrorAction SilentlyContinue)
  if ($yq) {
    $temp = New-TemporaryFile
    ($Swagger | ConvertTo-Json -Depth 100) | Out-File -FilePath $temp -Encoding UTF8
    & $yq eval -P $temp | Out-File -FilePath $Path -Encoding UTF8
    Remove-Item $temp -Force
  } else {
    Write-Host "yq not found; skipping YAML write"
  }
}

function Replace-PathParams {
  param([string]$Path)
  $replacements = @{
    '{id}' = '1'
    '{account_id}' = '1'
    '{vendor_id}' = '1'
    '{customer_id}' = '1'
    '{receipt_id}' = '1'
    '{address_id}' = '1'
  }
  $out = $Path
  foreach ($k in $replacements.Keys) {
    $out = $out -replace [regex]::Escape($k), $replacements[$k]
  }
  return $out
}

function Is-DangerousMethod {
  param([string]$Method,[string]$Path)
  $m = $Method.ToUpperInvariant()
  if ($m -in @('DELETE','PUT','PATCH')) { return $true }
  # Conservative: certain POST endpoints are safe
  $safePostPatterns = @(
    '/refresh$',
    '/preview',
    '/validate',
    '/validate-token$',
    '/health$',
    '/metrics$',
    '/analytics$',
    '/status$',
    '/summary$'
  )
  if ($m -eq 'POST') {
    foreach ($p in $safePostPatterns) { if ($Path -match $p) { return $false } }
    return $true
  }
  return $false
}

function Needs-Auth {
  param([object]$Op)
  if ($null -eq $Op) { return $false }
  $hasProp = $Op.PSObject.Properties.Name -contains 'security'
  if ($hasProp -and $null -ne $Op.security -and $Op.security.Count -gt 0) { return $true }
  return $false
}

function Test-Operation {
  param(
    [string]$BaseUrl,
    [string]$Path,
    [string]$Method,
    [object]$Op,
    [string]$Token
  )
  $fullPath = Replace-PathParams -Path $Path
  $uri = "$BaseUrl$fullPath"
  $headers = @{}
  $tryAuth = Needs-Auth -Op $Op
  if ($tryAuth -and $Token) { $headers['Authorization'] = "Bearer $Token" }

  $body = $null
  $contentType = 'application/json'
  if ($Method -eq 'POST' -or $Method -eq 'PUT' -or $Method -eq 'PATCH') {
    # If parameters include a body schema, send an empty object as placeholder
    $body = '{}' 
  }

  try {
    $resp = Invoke-WebRequest -Method $Method -Uri $uri -Headers $headers -ContentType $contentType -Body $body -UseBasicParsing -TimeoutSec 30
    return @{ status = $resp.StatusCode; ok = ($resp.StatusCode -ge 200 -and $resp.StatusCode -lt 300); uri = $uri }
  } catch {
    $err = $_.Exception
    $status = if ($err.Response -and $err.Response.StatusCode) { [int]$err.Response.StatusCode } else { 0 }
    # Retry with auth if unauthorized and we didn't try auth
    if ($status -eq 401 -and -not $tryAuth -and $Token) {
      try {
        $headers['Authorization'] = "Bearer $Token"
        $resp2 = Invoke-WebRequest -Method $Method -Uri $uri -Headers $headers -ContentType $contentType -Body $body -UseBasicParsing -TimeoutSec 30
        return @{ status = $resp2.StatusCode; ok = ($resp2.StatusCode -ge 200 -and $resp2.StatusCode -lt 300); uri = $uri }
      } catch {
        $err2 = $_.Exception
        $status2 = if ($err2.Response -and $err2.Response.StatusCode) { [int]$err2.Response.StatusCode } else { 0 }
        return @{ status = $status2; ok = $false; uri = $uri; error = $err2.Message }
      }
    }
    return @{ status = $status; ok = $false; uri = $uri; error = $err.Message }
  }
}

# Main

Write-Host "Loading Swagger JSON from $SwaggerJsonPath"
$swagger = Load-SwaggerJson -Path $SwaggerJsonPath

$token = $null
try {
  Write-Host "Authenticating as $AuthEmail..."
  $token = Get-AccessToken -BaseUrl $BaseUrl -Email $AuthEmail -Password $AuthPassword
  Write-Host "Got token (hidden)"
} catch {
  Write-Warning $_
  Write-Host "Attempting to register a temporary test user..."
  $reg = Try-RegisterAndGetToken -BaseUrl $BaseUrl
  if ($reg -and $reg.token) {
    $token = $reg.token
    Write-Host "Registered temp user ($($reg.email)). Using its token."
  } else {
    Write-Host "Proceeding without auth token. Secured endpoints will likely return 401."
  }
}

$results = @()
$paths = $swagger.paths.PSObject.Properties | ForEach-Object { $_.Name }

foreach ($path in $paths) {
  $opObj = $swagger.paths.$path
  $methodProps = $opObj.PSObject.Properties | Where-Object { $_.MemberType -eq 'NoteProperty' }
  foreach ($mp in $methodProps) {
    $method = $mp.Name.ToUpperInvariant()
    $op = $mp.Value

    # Skip dangerous operations (DELETE/PUT/PATCH and most POST)
    if (Is-DangerousMethod -Method $method -Path $path) {
      $results += [pscustomobject]@{ path=$path; method=$method; tested=$false; ok=$null; status=$null; uri=$null; reason='skipped-dangerous' }
      continue
    }

    $r = Test-Operation -BaseUrl $BaseUrl -Path $path -Method $method -Op $op -Token $token
    $errVal = $null
    if ($r -is [hashtable] -and $r.ContainsKey('error')) { $errVal = $r['error'] }
    $results += [pscustomobject]@{ path=$path; method=$method; tested=$true; ok=$r.ok; status=$r.status; uri=$r.uri; error=$errVal }
  }
}

# Summarize by path: keep if any method passed or if all were skipped-dangerous
$byPath = $results | Group-Object path
$failedPaths = @()
foreach ($g in $byPath) {
  $anyTested = @($g.Group | Where-Object { $_.tested -eq $true }).Count -gt 0
  $anyPass = @($g.Group | Where-Object { $_.ok -eq $true }).Count -gt 0
  if ($anyTested -and -not $anyPass) { $failedPaths += $g.Name }
}

# Save report
$report = @{ base_url = $BaseUrl; tested_at = (Get-Date); results = $results }
$report | ConvertTo-Json -Depth 8 | Out-File -FilePath $ReportPath -Encoding UTF8
Write-Host "Report saved: $ReportPath"

if ($Clean) {
  Write-Host "Cleaning Swagger docs: removing failed paths: $($failedPaths.Count)"
  if ($failedPaths.Count -gt 0) {
    $b1 = Backup-File -Path $SwaggerJsonPath
    foreach ($p in $failedPaths) {
      if ($swagger.paths.PSObject.Properties.Name -contains $p) {
        Write-Host "Removing $p from swagger.json"
        $swagger.paths.PSObject.Properties.Remove($p) | Out-Null
      }
    }
    Save-SwaggerJson -Swagger $swagger -Path $SwaggerJsonPath

    # YAML (optional, only if yq exists)
    if (Test-Path $SwaggerYamlPath) {
      try {
        $swaggerYaml = Load-SwaggerYaml -Path $SwaggerYamlPath
        if ($swaggerYaml -and $swaggerYaml.paths) {
          foreach ($p in $failedPaths) {
            if ($swaggerYaml.paths.PSObject.Properties.Name -contains $p) {
              Write-Host "Removing $p from swagger.yaml"
              $swaggerYaml.paths.PSObject.Properties.Remove($p) | Out-Null
            }
          }
          $b2 = Backup-File -Path $SwaggerYamlPath
          Save-SwaggerYaml -Swagger $swaggerYaml -Path $SwaggerYamlPath
        }
      } catch {
        Write-Warning "YAML cleanup skipped: $($_.Exception.Message)"
      }
    }
  }
  Write-Host "Cleanup complete."
} else {
  Write-Host "Dry run complete. No changes made to Swagger files. Use -Clean to remove failing paths."
}
