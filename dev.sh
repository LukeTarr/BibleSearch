#!/usr/bin/env bash
# Local dev server. Templates are read from disk (DEV=true in .env), so HTML edits show on refresh.
# Tailwind watches the templates and rebuilds assets/app.css; air rebuilds only when .go files change.
set -euo pipefail

if command -v tailwindcss >/dev/null; then
  tailwindcss -i assets/input.css -o assets/app.css --minify --watch &
  tailwind_pid=$!
  trap 'kill "$tailwind_pid"' EXIT
else
  echo "tailwindcss not found: assets/app.css won't be rebuilt when you add new classes" >&2
fi

if command -v air >/dev/null; then
  air --build.cmd "go build -o bin/BibleSearch ." --build.bin "bin/BibleSearch" --build.include_ext "go"
else
  go run .
fi
