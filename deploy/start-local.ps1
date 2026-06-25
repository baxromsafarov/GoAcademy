# GoAcademy — local run with the code sandbox + online judge enabled.
#
# The sandbox/judge execute untrusted Go in Docker, so they need the Go toolchain
# and Docker on the host — that is why the backend runs natively here (not in the
# minimal container image). Postgres runs in Docker.
#
# Usage (from anywhere):  powershell -ExecutionPolicy Bypass -File deploy\start-local.ps1
#
# Opens the backend (:8080) and frontend (:5173) in separate windows; close those
# windows to stop. Open http://localhost:5173 in a browser.

$ErrorActionPreference = "Stop"
$repo = (Resolve-Path "$PSScriptRoot\..").Path
$nodeDir = "C:\Program Files\nodejs"
if (Test-Path $nodeDir) { $env:Path = "$nodeDir;$env:Path" }

# ---- configuration (dev defaults; change JWT_SECRET / DB password for real use) ----
$env:DATABASE_URL          = "postgres://goacademy:goacademy_dev_password@localhost:5432/goacademy?sslmode=disable"
$env:JWT_SECRET            = "dev_only_change_me_secret_min_32_chars"
$env:APP_ENV               = "development"
$env:HTTP_PORT             = "8080"
$env:LOG_FORMAT            = "text"
$env:SANDBOX_ENABLED       = "true"
$env:SANDBOX_IMAGE         = "busybox"
$env:STORAGE_LOCAL_DIR     = "$repo\backend\storage"
$env:STORAGE_PUBLIC_BASE_URL = "/static"

Write-Host "1/5  Starting Postgres (Docker)..." -ForegroundColor Cyan
docker compose -f "$repo\deploy\docker-compose.yml" up -d postgres | Out-Null
do {
  Start-Sleep -Seconds 1
  $health = (docker inspect goacademy-postgres --format "{{.State.Health.Status}}" 2>$null)
} until ($health -eq "healthy")
Write-Host "     Postgres healthy."

Write-Host "2/5  Applying migrations..." -ForegroundColor Cyan
go -C "$repo\backend" run ./cmd/migrate up

Write-Host "3/5  Seeding content (4 languages)..." -ForegroundColor Cyan
go -C "$repo\backend" run ./cmd/seed

Write-Host "4/5  Pulling sandbox image (busybox)..." -ForegroundColor Cyan
docker image inspect busybox 2>$null | Out-Null
if ($LASTEXITCODE -ne 0) { docker pull busybox | Out-Null }

Write-Host "5/5  Building + launching backend (:8080) and frontend (:5173)..." -ForegroundColor Cyan
go -C "$repo\backend" build -o "$repo\backend\bin\goacademy-api.exe" ./cmd/api

# Backend in its own window (inherits the env above).
Start-Process powershell -ArgumentList "-NoExit", "-Command", "& '$repo\backend\bin\goacademy-api.exe'"

# Frontend dev server in its own window (Vite proxies /api -> :8080).
if (-not (Test-Path "$repo\frontend\node_modules")) {
  Push-Location "$repo\frontend"; npm ci; Pop-Location
}
Start-Process powershell -ArgumentList "-NoExit", "-Command", "`$env:Path='$nodeDir;'+`$env:Path; Set-Location '$repo\frontend'; npm run dev"

Write-Host ""
Write-Host "GoAcademy is starting:" -ForegroundColor Green
Write-Host "  Frontend:  http://localhost:5173"
Write-Host "  API:       http://localhost:8080/api/v1"
Write-Host "  (sandbox + online judge enabled; close the two windows to stop)"
