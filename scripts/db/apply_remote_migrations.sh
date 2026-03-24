#!/usr/bin/env bash
set -euo pipefail

DRY_RUN=0
if [[ "${1:-}" == "--dry-run" ]]; then
  DRY_RUN=1
fi

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "[db/migrate] DATABASE_URL is required." >&2
  exit 1
fi

if ! command -v psql >/dev/null 2>&1; then
  echo "[db/migrate] psql command not found in PATH." >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
MIGRATIONS_DIR="${REPO_ROOT}/server/migrations"

if [[ ! -d "${MIGRATIONS_DIR}" ]]; then
  echo "[db/migrate] Migrations directory not found: ${MIGRATIONS_DIR}" >&2
  exit 1
fi

shopt -s nullglob
MIGRATION_FILES=("${MIGRATIONS_DIR}"/*.up.sql)
shopt -u nullglob

if [[ ${#MIGRATION_FILES[@]} -eq 0 ]]; then
  echo "[db/migrate] No .up.sql migration files found in ${MIGRATIONS_DIR}" >&2
  exit 1
fi

IFS=$'\n' MIGRATION_FILES=($(printf '%s\n' "${MIGRATION_FILES[@]}" | sort))
unset IFS

echo "[db/migrate] Migrations directory: ${MIGRATIONS_DIR}"
echo "[db/migrate] Files discovered: ${#MIGRATION_FILES[@]}"

for file in "${MIGRATION_FILES[@]}"; do
  base_name="$(basename "${file}")"
  echo "[db/migrate] Applying ${base_name}"

  if [[ ${DRY_RUN} -eq 1 ]]; then
    continue
  fi

  psql "${DATABASE_URL}" -v ON_ERROR_STOP=1 -f "${file}"
done

if [[ ${DRY_RUN} -eq 1 ]]; then
  echo "[db/migrate] Dry-run complete; no migrations were executed."
else
  echo "[db/migrate] All migrations applied successfully."
fi
