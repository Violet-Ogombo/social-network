# Smoke test script for the social-network app
# Usage: run this while the backend is running (or adjust to start it here)

$ErrorActionPreference = 'Stop'

$base = 'http://localhost:8080'

# simple register
$body = @{
  email = "smoketest@example.com"
  password = "Sm0keTest!"
  first_name = "Smoke"
  last_name = "Tester"
} | ConvertTo-Json

Write-Host "Registering user..."
try {
  $res = Invoke-RestMethod -Uri "$base/register" -Method Post -Body $body -ContentType 'application/json' -SkipCertificateCheck
  Write-Host "Register response:" ($res | ConvertTo-Json -Depth 5)
} catch {
  Write-Host "Register failed:" $_.Exception.Response.StatusCode
  $_.Exception.Response.GetResponseStream() | % { [Console]::Out.WriteLine((New-Object System.IO.StreamReader($_)).ReadToEnd()) }
  exit 1
}

Write-Host "Checking session..."
try {
  $res = Invoke-RestMethod -Uri "$base/api/check-session" -Method Get -UseBasicParsing -Credential $null -SkipCertificateCheck -Headers @{Cookie=$(Get-ChildItem Variable: | Where-Object { $_.Name -eq 'COOKIE' } )}
  Write-Host "Check session response:" ($res | ConvertTo-Json -Depth 5)
} catch {
  Write-Host "Check session failed:" $_
}

Write-Host "Smoke test completed. Please manually verify WebSocket behavior via the SPA or dev UI."
