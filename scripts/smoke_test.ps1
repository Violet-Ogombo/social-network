# Smoke test script for the social-network app
# Usage: run this while the backend is running (or adjust to start it here)

$ErrorActionPreference = 'Stop'

$base = 'http://localhost:8080'

# Use a WebRequestSession so cookies are preserved across calls (works in Windows PowerShell and PowerShell Core)
$session = New-Object Microsoft.PowerShell.Commands.WebRequestSession

# simple register
$body = @{
  email = "smoketest@example.com"
  password = "Sm0keTest!"
  first_name = "Smoke"
  last_name = "Tester"
} | ConvertTo-Json

Write-Host "Registering user..."
try {
  $res = Invoke-RestMethod -Uri "$base/register" -Method Post -Body $body -ContentType 'application/json' -WebSession $session
  Write-Host "Register response:" ($res | ConvertTo-Json -Depth 5)
} catch {
  Write-Host "Register failed:" $_.Exception.Message
  if ($_.Exception -and $_.Exception.Response) {
    try {
      $stream = $_.Exception.Response.GetResponseStream()
      $reader = New-Object System.IO.StreamReader($stream)
      $bodyText = $reader.ReadToEnd()
      Write-Host "Response body:" $bodyText
    } catch {
      Write-Host "Unable to read response stream:" $_
    }
  }
  exit 1
}

Write-Host "Checking session..."
try {
  $res = Invoke-RestMethod -Uri "$base/api/check-session" -Method Get -WebSession $session
  Write-Host "Check session response:" ($res | ConvertTo-Json -Depth 5)
} catch {
  Write-Host "Check session failed:" $_.Exception.Message
}

Write-Host "Smoke test completed. Please manually verify WebSocket behavior via the SPA or dev UI."
