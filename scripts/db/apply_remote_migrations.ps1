param(
  [switch]$DryRun
)

$ErrorActionPreference = 'Stop'

function Require-EnvVar {
  param([string]$Name)

  if ([string]::IsNullOrWhiteSpace((Get-Item -Path "Env:$Name" -ErrorAction SilentlyContinue).Value)) {
    throw "Environment variable '$Name' is required."
  }
}

function Require-Command {
  param([string]$Name)

  if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
    throw "Required command '$Name' was not found in PATH."
  }
}

Require-EnvVar -Name 'DATABASE_URL'
Require-Command -Name 'psql'

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
$migrationsDir = Join-Path $repoRoot 'server\migrations'

if (-not (Test-Path -Path $migrationsDir -PathType Container)) {
  throw "Migrations directory was not found: $migrationsDir"
}

$migrationFiles = Get-ChildItem -Path $migrationsDir -Filter '*.up.sql' -File | Sort-Object Name

if ($migrationFiles.Count -eq 0) {
  throw "No .up.sql migration files were found in $migrationsDir"
}

Write-Host "[db/migrate] Migrations directory: $migrationsDir"
Write-Host "[db/migrate] Files discovered: $($migrationFiles.Count)"

foreach ($file in $migrationFiles) {
  Write-Host "[db/migrate] Applying $($file.Name)"

  if ($DryRun) {
    continue
  }

  & psql $env:DATABASE_URL -v ON_ERROR_STOP=1 -f $file.FullName | Out-Host
  if ($LASTEXITCODE -ne 0) {
    throw "Migration failed: $($file.Name)"
  }
}

if ($DryRun) {
  Write-Host '[db/migrate] Dry-run complete; no migrations were executed.'
} else {
  Write-Host '[db/migrate] All migrations applied successfully.'
}
