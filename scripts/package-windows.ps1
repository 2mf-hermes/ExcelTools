# Package Windows artifacts (portable + optional NSIS)
# Usage: powershell -File scripts/package-windows.ps1

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

Write-Host "== Tests =="
go test ./...
if ($LASTEXITCODE -ne 0) { throw "go test failed" }

Write-Host "== i18n =="
Push-Location frontend
npm run test:i18n
if ($LASTEXITCODE -ne 0) { Pop-Location; throw "i18n check failed" }
Pop-Location

Write-Host "== Wails build (portable) =="
wails build
if ($LASTEXITCODE -ne 0) { throw "wails build failed" }

$dist = Join-Path $root "dist"
New-Item -ItemType Directory -Force -Path $dist | Out-Null

$exe = Join-Path $root "build/bin/ExcelTools.exe"
if (!(Test-Path $exe)) { throw "missing $exe" }

$zip = Join-Path $dist "ExcelTools-portable-win64.zip"
if (Test-Path $zip) { Remove-Item $zip }
Compress-Archive -Path $exe -DestinationPath $zip
Write-Host "Portable zip: $zip"

# Optional NSIS installer
$makensis = Get-Command makensis -ErrorAction SilentlyContinue
if ($makensis) {
  Write-Host "== NSIS installer =="
  wails build -nsis
  Get-ChildItem build/bin/*-installer.exe -ErrorAction SilentlyContinue | ForEach-Object {
    Copy-Item $_.FullName -Destination $dist -Force
    Write-Host "Installer: $($_.Name)"
  }
} else {
  Write-Host "makensis not found — skipping installer (portable zip only)"
}

Write-Host "Done. Artifacts in $dist"
Get-ChildItem $dist
