#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

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

read_package_name() {
  local target_dir="$1"
  if [[ ! -f "${target_dir}/package.json" ]]; then
    return
  fi
  if command -v node >/dev/null 2>&1; then
    node -e "const pkg=require('${target_dir}/package.json'); if (pkg.name) console.log(pkg.name)" 2>/dev/null || true
    return
  fi
  grep -E '"name"\s*:' "${target_dir}/package.json" | head -n 1 | sed -E 's/.*"name"\s*:\s*"([^"]+)".*/\1/' || true
}

read_module_name() {
  local target_dir="$1"
  if [[ ! -f "${target_dir}/go.mod" ]]; then
    return
  fi
  awk '/^module / { print $2; exit }' "${target_dir}/go.mod" 2>/dev/null || true
}

read_pm_version() {
  local pm="$1"
  if ! command -v "${pm}" >/dev/null 2>&1; then
    printf '%s: not installed' "${pm}"
    return
  fi
  case "${pm}" in
    pnpm)
      pnpm --version 2>/dev/null | sed 's/^/pnpm /'
      ;;
    yarn)
      yarn --version 2>/dev/null | sed 's/^/yarn /'
      ;;
    npm)
      npm --version 2>/dev/null | sed 's/^/npm /'
      ;;
  esac
}

cd "${ROOT_DIR}"

frontend_dir="$(detect_frontend_dir || true)"
backend_dir="$(detect_backend_dir || true)"
package_name=""
package_manager=""
module_name=""
pm_version="frontend package manager: not detected"

if [[ -n "${frontend_dir}" ]]; then
  package_name="$(read_package_name "${frontend_dir}")"
  package_manager="$(detect_package_manager "${frontend_dir}")"
  pm_version="$(read_pm_version "${package_manager}")"
fi

if [[ -n "${backend_dir}" ]]; then
  module_name="$(read_module_name "${backend_dir}")"
fi

node_version="$(node --version 2>/dev/null || echo 'node: not installed')"
go_version="$(go version 2>/dev/null || echo 'go: not installed')"
docker_version="$(
  (docker compose version 2>/dev/null || docker --version 2>/dev/null) | head -n 1 || true
)"

if [[ -z "${docker_version}" ]]; then
  docker_version="docker: not installed"
fi

echo "== SteelLink Admin Harness Bootstrap =="
echo "root: ${ROOT_DIR}"
if [[ -n "${frontend_dir}" ]]; then
  echo "frontend dir: ${frontend_dir}"
else
  echo "frontend dir: not detected"
fi
if [[ -n "${package_name}" ]]; then
  echo "frontend package: ${package_name}"
else
  echo "frontend package: package.json not found"
fi
if [[ -n "${backend_dir}" ]]; then
  echo "backend dir: ${backend_dir}"
else
  echo "backend dir: not detected"
fi
if [[ -n "${module_name}" ]]; then
  echo "backend module: ${module_name}"
else
  echo "backend module: go.mod not found"
fi
echo "${node_version}"
echo "${go_version}"
echo "${pm_version}"
echo "${docker_version}"

if git rev-parse --show-toplevel >/dev/null 2>&1; then
  branch="$(git symbolic-ref --quiet --short HEAD 2>/dev/null || echo 'DETACHED')"
  remote_url="$(git remote get-url origin 2>/dev/null || echo 'origin: not configured')"
  echo "git branch: ${branch}"
  echo "git remote: ${remote_url}"
else
  echo "git: current directory is not an initialized repository"
fi

echo
echo "Read in order:"
echo "1. AGENTS.md"
echo "2. docs/ai/entrypoint.md"
echo "3. docs/ai/project-map.md"
echo "4. docs/ai/engineering-rules.md"
echo "5. docs/ai/git-workflow.md"
echo "6. docs/ai/review-rules.md"
echo "7. docs/ai/release-rules.md"
echo
echo "Key commands:"
echo "- ./harness/branch-guard.sh local"
echo "- ./harness/verify.sh"
echo "- ./harness/review.sh docs/features/<feature-id>"
echo "- ./harness/release.sh docs/features/<feature-id>"
echo "- ./harness/deploy.sh staging"
echo
echo "Notes:"
echo "- verify can validate frontend package.json scripts and backend go.mod in one pass"
echo "- if your monorepo does not use root/frontend/backend defaults, set FRONTEND_DIR or BACKEND_DIR"
echo "- branch-guard requires a valid Git worktree and current repository remote"
