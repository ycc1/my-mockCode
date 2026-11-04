$ErrorActionPreference = "Stop"

Set-Location $PSScriptRoot

$envFile = Join-Path $PSScriptRoot ".env"
if (-not (Test-Path $envFile)) {
    Copy-Item (Join-Path $PSScriptRoot ".env.example") $envFile
    Write-Host "Created .env from .env.example. Review it and add the UAT private-key path."
}

npm ci
if ($LASTEXITCODE -ne 0) {
    throw "Could not install the Unico Web SDK dependencies."
}
npm run build
if ($LASTEXITCODE -ne 0) {
    throw "Could not build the Unico Web SDK browser bundle."
}

go run .
