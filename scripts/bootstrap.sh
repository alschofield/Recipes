#!/usr/bin/env bash
set -euo pipefail

echo "[bootstrap] Checking developer tooling..."

have() {
  command -v "$1" >/dev/null 2>&1
}

install_hint() {
  local tool="$1"
  local hint="$2"
  echo "[bootstrap] Missing ${tool}. Install with: ${hint}"
}

if have task; then
  echo "[bootstrap] task: ok ($(task --version 2>/dev/null || echo installed))"
  HAVE_TASK=1
else
  HAVE_TASK=0
  if have brew; then
    echo "[bootstrap] Installing task via brew..."
    brew install go-task/tap/go-task
  elif have apt-get; then
    echo "[bootstrap] Installing task via apt..."
    sudo apt-get update && sudo apt-get install -y task
  elif have pacman; then
    echo "[bootstrap] Installing task via pacman..."
    sudo pacman -S --noconfirm go-task
  else
    install_hint "task" "https://taskfile.dev/installation/"
  fi
fi

if have make; then
  echo "[bootstrap] make: ok ($(make --version | head -n 1))"
  HAVE_MAKE=1
else
  HAVE_MAKE=0
  if have brew; then
    echo "[bootstrap] make missing (optional). Install with: brew install make"
  elif have apt-get; then
    echo "[bootstrap] make missing (optional). Install with: sudo apt-get install -y build-essential"
  else
    echo "[bootstrap] make missing (optional). You can use task instead."
  fi
fi

if have pnpm; then
  echo "[bootstrap] pnpm: ok ($(pnpm --version))"
else
  echo "[bootstrap] pnpm missing. Install with: corepack enable && corepack prepare pnpm@10 --activate"
fi

if have go; then
  echo "[bootstrap] go: ok ($(go version))"
else
  echo "[bootstrap] go missing. Install Go 1.25+ from https://go.dev/dl/"
fi

if have docker; then
  echo "[bootstrap] docker: ok ($(docker --version))"
else
  echo "[bootstrap] docker missing. Install Docker Desktop / Engine"
fi

echo "[bootstrap] Done. Next: task setup (or make setup)"

if [[ "${HAVE_TASK}" -eq 0 && "${HAVE_MAKE}" -eq 0 ]]; then
  echo "[bootstrap] Neither task nor make is available."
  echo "[bootstrap] Migration fallback command:"
  echo "[bootstrap] MSYS_NO_PATHCONV=1 docker run --rm -v \"$(pwd)/server/migrations:/migrations\" migrate/migrate -path=/migrations -database \"postgres://postgres:postgres@localhost:5432/recipes?sslmode=disable\" up"
fi
