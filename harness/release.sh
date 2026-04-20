#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FEATURE_DIR="${1:-}"

if [[ -z "${FEATURE_DIR}" ]]; then
  FEATURE_DIR="$("${ROOT_DIR}/harness/resolve-feature-dir.sh")"
fi

cd "${ROOT_DIR}"

for file in spec.md test-cases.md implementation-plan.md release-checklist.md; do
  if [[ ! -f "${FEATURE_DIR}/${file}" ]]; then
    echo "missing required release artifact: ${FEATURE_DIR}/${file}" >&2
    exit 1
  fi
done

frontend_detected=0
backend_detected=0

if [[ -f "package.json" || -f "frontend/package.json" || -f "web/package.json" || -f "ui/package.json" || -f "client/package.json" || -f "admin-web/package.json" ]]; then
  frontend_detected=1
fi

if [[ -f "go.mod" || -f "backend/go.mod" || -f "server/go.mod" || -f "api/go.mod" ]]; then
  backend_detected=1
fi

if [[ "${frontend_detected}" -eq 0 && "${backend_detected}" -eq 0 ]]; then
  echo "missing required release file: neither frontend package.json nor backend go.mod was detected" >&2
  exit 1
fi

echo "== Release Gate =="
echo "feature: ${FEATURE_DIR}"
./harness/verify.sh

if compgen -G "Dockerfile*" >/dev/null 2>&1; then
  echo "dockerfile present"
else
  echo "dockerfile not found; ensure deployment flow is documented in release checklist"
fi

if compgen -G ".env*.example" >/dev/null 2>&1; then
  echo "environment example file present"
else
  echo "environment example file not found; ensure required env vars are documented"
fi

echo
echo "release gate passed"
