# GoAcademy — share/present from anywhere.
#
# Starts the full stack (Postgres in Docker, native backend with the code
# sandbox, Vite dev server bound to all interfaces) and a free cloudflared
# "quick tunnel", then prints the public HTTPS link.
#
# Run this in YOUR OWN terminal so it keeps running independently:
#   powershell -ExecutionPolicy Bypass -File deploy\start-share.ps1
#
# Each window stays open; close them to stop. Keep the PC awake and on AC power.
# The tunnel URL changes every time cloudflared restarts (free quick tunnels).

$ErrorActionPreference = "Stop"
$repo = (Resolve-Path "$PSScriptRoot\..").Path
$nodeDir = "C:\Program Files\nodejs"
if (Test-Path $nodeDir) { $env:Path = "$nodeDir;$env:Path" }

# ---- backend configuration (dev defaults) ----
$env:DATABASE_URL            = "postgres://goacademy:goacademy_dev_password@localhost:5432/goacademy?sslmode=disable"
$env:JWT_SECRET              = "dev_only_change_me_secret_min_32_chars"
$env:APP_ENV                 = "development"
$env:HTTP_PORT               = "8080"
$env:LOG_FORMAT              = "text"
$env:SANDBOX_ENABLED         = "true"
$env:SANDBOX_IMAGE           = "busybox"
$env:STORAGE_LOCAL_DIR       = "$repo\backend\storage"
$env:STORAGE_PUBLIC_BASE_URL = "/static"

Write-Host "1/4  Starting Postgres (Docker)..." -ForegroundColor Cyan
docker compose -f "$repo\deploy\docker-compose.yml" up -d postgres | Out-Null
do {
  Start-Sleep -Seconds 1
  $health = (docker inspect goacademy-postgres --format "{{.State.Health.Status}}" 2>$null)
} until ($health -eq "healthy")
Write-Host "     Postgres healthy."

Write-Host "2/4  Applying migrations + building backend..." -ForegroundColor Cyan
go -C "$repo\backend" run ./cmd/migrate up
go -C "$repo\backend" build -o "$repo\backend\bin\goacademy-api.exe" ./cmd/api

Write-Host "3/4  Launching backend (:8080) and frontend (:5173)..." -ForegroundColor Cyan
# Backend in its own window (inherits the env above).
Start-Process powershell -ArgumentList "-NoExit", "-Command", "& '$repo\backend\bin\goacademy-api.exe'"
# Frontend dev server (vite.config has host:true + allowedHosts:true).
if (-not (Test-Path "$repo\frontend\node_modules")) {
  Push-Location "$repo\frontend"; npm ci; Pop-Location
}
Start-Process powershell -ArgumentList "-NoExit", "-Command", "`$env:Path='$nodeDir;'+`$env:Path; Set-Location '$repo\frontend'; npm run dev"

# Wait for Vite to answer before opening the tunnel.
$ready = $false
for ($i = 0; $i -lt 30; $i++) {
  Start-Sleep -Seconds 1
  try { Invoke-WebRequest -UseBasicParsing -TimeoutSec 2 "http://localhost:5173/" | Out-Null; $ready = $true; break } catch {}
}
if (-not $ready) { Write-Host "     (frontend still starting; tunnel will retry)" -ForegroundColor Yellow }

Write-Host "4/4  Opening the public tunnel (cloudflared)..." -ForegroundColor Cyan
$cf = (Get-Command cloudflared -ErrorAction SilentlyContinue).Source
if (-not $cf) { $cf = "C:\Program Files (x86)\cloudflared\cloudflared.exe" }
if (-not (Test-Path $cf)) {
  Write-Host "cloudflared not found. Install it:  winget install Cloudflare.cloudflared" -ForegroundColor Red
  exit 1
}
$log = "$env:TEMP\goacademy-tunnel.log"
Remove-Item $log -ErrorAction SilentlyContinue
Start-Process -FilePath $cf -ArgumentList "tunnel", "--url", "http://localhost:5173", "--no-autoupdate" `
  -RedirectStandardError $log -WindowStyle Minimized

$url = $null
for ($i = 0; $i -lt 40; $i++) {
  Start-Sleep -Seconds 1
  if (Test-Path $log) {
    $m = Select-String -Path $log -Pattern 'https://[a-z0-9-]+\.trycloudflare\.com' -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($m) { $url = $m.Matches[0].Value; break }
  }
}

Write-Host ""
if ($url) {
  Write-Host "================================================================" -ForegroundColor Green
  Write-Host "  GoAcademy is live. Open this link from any device:" -ForegroundColor Green
  Write-Host "      $url" -ForegroundColor White
  Write-Host "================================================================" -ForegroundColor Green
} else {
  Write-Host "Tunnel didn't report a URL yet — check $log" -ForegroundColor Yellow
}
Write-Host ""
Write-Host "Local:  http://localhost:5173    API: http://localhost:8080/api/v1"
Write-Host "Keep this and the backend/frontend windows open. Close them to stop."
