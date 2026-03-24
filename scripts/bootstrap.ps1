Write-Host "[bootstrap] Checking developer tooling..."

function Has-Cmd($name) {
  return $null -ne (Get-Command $name -ErrorAction SilentlyContinue)
}

if (Has-Cmd task) {
  Write-Host "[bootstrap] task: ok"
  $haveTask = $true
} else {
  $haveTask = $false
  if (Has-Cmd winget) {
    Write-Host "[bootstrap] Installing task via winget..."
    winget install --id Task.Task --accept-source-agreements --accept-package-agreements
  } else {
    Write-Host "[bootstrap] Missing task. Install from https://taskfile.dev/installation/"
  }
}

if (Has-Cmd make) {
  Write-Host "[bootstrap] make: ok"
  $haveMake = $true
} else {
  $haveMake = $false
  Write-Host "[bootstrap] make missing (optional). Task is the primary cross-platform runner."
}

if (Has-Cmd pnpm) {
  Write-Host "[bootstrap] pnpm: ok"
} else {
  Write-Host "[bootstrap] pnpm missing. Run: corepack enable; corepack prepare pnpm@10 --activate"
}

if (Has-Cmd go) {
  Write-Host "[bootstrap] go: ok"
} else {
  Write-Host "[bootstrap] go missing. Install Go 1.25+ from https://go.dev/dl/"
}

if (Has-Cmd docker) {
  Write-Host "[bootstrap] docker: ok"
} else {
  Write-Host "[bootstrap] docker missing. Install Docker Desktop"
}

Write-Host "[bootstrap] Done. Next: task setup (or make setup)"

if (-not $haveTask -and -not $haveMake) {
  Write-Host "[bootstrap] Neither task nor make is available."
  Write-Host "[bootstrap] Migration fallback command:"
  Write-Host "[bootstrap] MSYS_NO_PATHCONV=1 docker run --rm -v `"$PWD/server/migrations:/migrations`" migrate/migrate -path=/migrations -database `"postgres://postgres:postgres@localhost:5432/recipes?sslmode=disable`" up"
}
