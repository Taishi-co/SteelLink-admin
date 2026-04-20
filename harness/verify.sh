#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

fail() {
  echo "verify failed: $1" >&2
  exit 1
}

detect_frontend_dir() {
  if [[ -n "${FRONTEND_DIR:-}" && -f "${ROOT_DIR}/${FRONTEND_DIR}/package.json" ]]; then
    printf '%s' "${ROOT_DIR}/${FRONTEND_DIR}"
    return
  fi
  if [[ -f "${ROOT_DIR}/package.json" ]]; then
    printf '%s' "${ROOT_DIR}"
    return
  fi
  for dir in frontend web ui client admin-web; do
    if [[ -f "${ROOT_DIR}/${dir}/package.json" ]]; then
      printf '%s' "${ROOT_DIR}/${dir}"
      return
    fi
  done
}

detect_backend_dir() {
  if [[ -n "${BACKEND_DIR:-}" && -f "${ROOT_DIR}/${BACKEND_DIR}/go.mod" ]]; then
    printf '%s' "${ROOT_DIR}/${BACKEND_DIR}"
    return
  fi
  if [[ -f "${ROOT_DIR}/go.mod" ]]; then
    printf '%s' "${ROOT_DIR}"
    return
  fi
  for dir in backend server api; do
    if [[ -f "${ROOT_DIR}/${dir}/go.mod" ]]; then
      printf '%s' "${ROOT_DIR}/${dir}"
      return
    fi
  done
}

detect_package_manager() {
  local target_dir="$1"
  if [[ -f "${target_dir}/pnpm-lock.yaml" ]]; then
    printf 'pnpm'
    return
  fi
  if [[ -f "${target_dir}/yarn.lock" ]]; then
    printf 'yarn'
    return
  fi
  if [[ -f "${target_dir}/package-lock.json" ]]; then
    printf 'npm'
    return
  fi
  printf 'npm'
}

has_script() {
  local target_dir="$1"
  local script_name="$2"
  node -e "const pkg=require('${target_dir}/package.json'); const scripts=pkg.scripts || {}; process.exit(Object.prototype.hasOwnProperty.call(scripts, '${script_name}') ? 0 : 1)"
}

run_script() {
  local target_dir="$1"
  local script_name="$2"
  case "${PACKAGE_MANAGER}" in
    pnpm)
      (cd "${target_dir}" && pnpm run "${script_name}")
      ;;
    yarn)
      (cd "${target_dir}" && yarn "${script_name}")
      ;;
    npm)
      (cd "${target_dir}" && npm run "${script_name}")
      ;;
    *)
      fail "unsupported package manager: ${PACKAGE_MANAGER}"
      ;;
  esac
}

run_script_with_ci() {
  local target_dir="$1"
  local script_name="$2"
  case "${PACKAGE_MANAGER}" in
    pnpm)
      (cd "${target_dir}" && CI=1 pnpm run "${script_name}")
      ;;
    yarn)
      (cd "${target_dir}" && CI=1 yarn "${script_name}")
      ;;
    npm)
      (cd "${target_dir}" && CI=1 npm run "${script_name}")
      ;;
    *)
      fail "unsupported package manager: ${PACKAGE_MANAGER}"
      ;;
  esac
}

detect_go_build_target() {
  local target_dir="$1"
  if [[ -n "${BACKEND_BUILD_TARGET:-}" ]]; then
    printf '%s' "${BACKEND_BUILD_TARGET}"
    return
  fi
  if [[ -d "${target_dir}/cmd/api" ]]; then
    printf '%s' "./cmd/api"
    return
  fi
  if [[ -d "${target_dir}/cmd/server" ]]; then
    printf '%s' "./cmd/server"
    return
  fi
  if [[ -d "${target_dir}/cmd" ]]; then
    printf '%s' "./cmd/..."
    return
  fi
  printf '%s' "./..."
}

cd "${ROOT_DIR}"

FRONTEND_TARGET_DIR="$(detect_frontend_dir || true)"
BACKEND_TARGET_DIR="$(detect_backend_dir || true)"

if [[ -z "${FRONTEND_TARGET_DIR}" && -z "${BACKEND_TARGET_DIR}" ]]; then
  fail "neither frontend package.json nor backend go.mod was detected; initialize the monorepo before running verify"
fi

echo "== Verify =="

if [[ -n "${FRONTEND_TARGET_DIR}" ]]; then
  if ! command -v node >/dev/null 2>&1; then
    fail "node is not installed"
  fi

  PACKAGE_MANAGER="$(detect_package_manager "${FRONTEND_TARGET_DIR}")"

  if ! command -v "${PACKAGE_MANAGER}" >/dev/null 2>&1; then
    fail "${PACKAGE_MANAGER} is not installed"
  fi

  if ! has_script "${FRONTEND_TARGET_DIR}" build; then
    fail "frontend package.json is missing required script: build"
  fi

  echo "[1/6] frontend lint (optional)"
  if has_script "${FRONTEND_TARGET_DIR}" lint; then
    run_script "${FRONTEND_TARGET_DIR}" lint
  else
    echo "frontend lint script not configured, skipped"
  fi
  echo

  echo "[2/6] frontend typecheck (optional)"
  if has_script "${FRONTEND_TARGET_DIR}" typecheck; then
    run_script "${FRONTEND_TARGET_DIR}" typecheck
  else
    echo "frontend typecheck script not configured, skipped"
  fi
  echo

  echo "[3/6] frontend test (optional)"
  if has_script "${FRONTEND_TARGET_DIR}" test:ci; then
    run_script_with_ci "${FRONTEND_TARGET_DIR}" test:ci
  elif has_script "${FRONTEND_TARGET_DIR}" test; then
    run_script_with_ci "${FRONTEND_TARGET_DIR}" test
  else
    echo "frontend test script not configured, skipped"
  fi
  echo

  echo "[4/6] frontend build"
  run_script "${FRONTEND_TARGET_DIR}" build
  echo
else
  echo "[1/6] frontend checks skipped: package.json not detected"
  echo
  echo "[2/6] frontend checks skipped: package.json not detected"
  echo
  echo "[3/6] frontend checks skipped: package.json not detected"
  echo
  echo "[4/6] frontend checks skipped: package.json not detected"
  echo
fi

if [[ -n "${BACKEND_TARGET_DIR}" ]]; then
  if ! command -v go >/dev/null 2>&1; then
    fail "go is not installed"
  fi

  BACKEND_BUILD_TARGET_DETECTED="$(detect_go_build_target "${BACKEND_TARGET_DIR}")"

  echo "[5/6] backend test"
  (cd "${BACKEND_TARGET_DIR}" && go test ./...)
  echo

  echo "[6/6] backend build"
  (cd "${BACKEND_TARGET_DIR}" && go build "${BACKEND_BUILD_TARGET_DETECTED}")
  echo
else
  echo "[5/6] backend checks skipped: go.mod not detected"
  echo
  echo "[6/6] backend checks skipped: go.mod not detected"
  echo
fi

echo "[extra] docker compose config (optional)"
compose_detected=0
for candidate in docker-compose.yml compose.yml compose.yaml; do
  if [[ -f "${candidate}" ]]; then
    compose_detected=1
    break
  fi
done

if [[ "${compose_detected}" -eq 1 ]]; then
  if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    docker compose config >/dev/null
    echo "docker compose config passed"
  else
    echo "compose file exists but docker compose is not available locally, skipped"
  fi
else
  echo "compose file not found, skipped"
fi
echo

echo "verify passed"
