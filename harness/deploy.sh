#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TARGET_ENV="${1:-}"
ENV_FILE="${ENV_FILE:-.env.${TARGET_ENV}}"
DEPLOY_SERVICE="${DEPLOY_SERVICE:-}"
STATIC_OUT_DIR="${STATIC_OUT_DIR:-dist}"
ADMIN_BACKEND_DOCKERFILE="${ADMIN_BACKEND_DOCKERFILE:-Dockerfile.admin-backend}"
ADMIN_FRONTEND_DOCKERFILE="${ADMIN_FRONTEND_DOCKERFILE:-Dockerfile.admin-frontend}"
ADMIN_BACKEND_IMAGE="${ADMIN_BACKEND_IMAGE:-steellink-admin-backend:${TARGET_ENV}}"
ADMIN_FRONTEND_IMAGE="${ADMIN_FRONTEND_IMAGE:-steellink-admin-frontend:${TARGET_ENV}}"
ADMIN_FRONTEND_API_BASE_URL="${ADMIN_FRONTEND_API_BASE_URL:-}"

if [[ "${TARGET_ENV}" != "staging" && "${TARGET_ENV}" != "prod" ]]; then
  echo "usage: ./harness/deploy.sh <staging|prod>" >&2
  exit 1
fi

cd "${ROOT_DIR}"

compose_file=""
for candidate in docker-compose.yml compose.yml compose.yaml; do
  if [[ -f "${candidate}" ]]; then
    compose_file="${candidate}"
    break
  fi
done

echo "== Deploy Plan =="
echo "target environment: ${TARGET_ENV}"

if [[ -d "${STATIC_OUT_DIR}" ]]; then
  echo "static output detected: ${STATIC_OUT_DIR}"
else
  echo "static output not found: ${STATIC_OUT_DIR} (build before deployment if your platform requires it)"
fi

planned_commands=()
deployment_mode=""

if [[ -n "${DEPLOY_COMMAND:-}" ]]; then
  planned_commands=("${DEPLOY_COMMAND}")
  deployment_mode="custom"
  echo "deployment entrypoint: DEPLOY_COMMAND"
elif [[ -f "${ADMIN_BACKEND_DOCKERFILE}" && -f "${ADMIN_FRONTEND_DOCKERFILE}" ]]; then
  deployment_mode="admin-dockerfiles"
  planned_commands+=("docker build -f ${ADMIN_BACKEND_DOCKERFILE} -t ${ADMIN_BACKEND_IMAGE} .")
  frontend_build_command="docker build -f ${ADMIN_FRONTEND_DOCKERFILE}"
  if [[ -n "${ADMIN_FRONTEND_API_BASE_URL}" ]]; then
    frontend_build_command+=" --build-arg VITE_API_BASE_URL=${ADMIN_FRONTEND_API_BASE_URL}"
  fi
  frontend_build_command+=" -t ${ADMIN_FRONTEND_IMAGE} ."
  planned_commands+=("${frontend_build_command}")
  echo "deployment entrypoint: admin backend/frontend Dockerfiles"
elif [[ -n "${compose_file}" && -f "Dockerfile" ]]; then
  deployment_mode="compose"
  if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    if [[ -f "${ENV_FILE}" ]]; then
      docker compose --env-file "${ENV_FILE}" config >/dev/null
      echo "validated compose plan with ${ENV_FILE}"
      if [[ -n "${DEPLOY_SERVICE}" ]]; then
        planned_commands+=("docker compose --env-file ${ENV_FILE} up -d --build ${DEPLOY_SERVICE}")
      else
        planned_commands+=("docker compose --env-file ${ENV_FILE} up -d --build")
      fi
    else
      docker compose config >/dev/null
      echo "validated compose plan with default environment"
      if [[ -n "${DEPLOY_SERVICE}" ]]; then
        planned_commands+=("docker compose up -d --build ${DEPLOY_SERVICE}")
      else
        planned_commands+=("docker compose up -d --build")
      fi
    fi
  else
    echo "docker compose file exists but docker compose is not available locally"
    if [[ -n "${DEPLOY_SERVICE}" ]]; then
      planned_commands+=("docker compose up -d --build ${DEPLOY_SERVICE}")
    else
      planned_commands+=("docker compose up -d --build")
    fi
  fi
elif [[ -f "Dockerfile" ]]; then
  deployment_mode="dockerfile"
  planned_commands+=("docker build -t steellink-admin:${TARGET_ENV} .")
else
  echo "no deploy entrypoint detected; set DEPLOY_COMMAND or add Dockerfile/Compose config"
fi

if [[ "${#planned_commands[@]}" -gt 0 ]]; then
  for command in "${planned_commands[@]}"; do
    echo "planned command: ${command}"
  done
fi

if [[ "${DEPLOY_EXECUTE:-0}" != "1" ]]; then
  echo "dry-run mode"
  echo "set DEPLOY_EXECUTE=1 to enable real deployment steps"
  exit 0
fi

if [[ "${#planned_commands[@]}" -eq 0 ]]; then
  echo "deployment command is not configured" >&2
  exit 1
fi

if [[ "${deployment_mode}" == "custom" ]]; then
  bash -lc "${DEPLOY_COMMAND}"
elif [[ "${deployment_mode}" == "admin-dockerfiles" ]]; then
  if ! command -v docker >/dev/null 2>&1; then
    echo "docker is required when DEPLOY_EXECUTE=1 and admin Dockerfile deployment is selected" >&2
    exit 1
  fi
  docker build -f "${ADMIN_BACKEND_DOCKERFILE}" -t "${ADMIN_BACKEND_IMAGE}" .
  frontend_build=(docker build -f "${ADMIN_FRONTEND_DOCKERFILE}")
  if [[ -n "${ADMIN_FRONTEND_API_BASE_URL}" ]]; then
    frontend_build+=(--build-arg "VITE_API_BASE_URL=${ADMIN_FRONTEND_API_BASE_URL}")
  fi
  frontend_build+=(-t "${ADMIN_FRONTEND_IMAGE}" .)
  "${frontend_build[@]}"
elif [[ "${deployment_mode}" == "compose" ]]; then
  if ! command -v docker >/dev/null 2>&1 || ! docker compose version >/dev/null 2>&1; then
    echo "docker compose is required when DEPLOY_EXECUTE=1 and compose deployment is selected" >&2
    exit 1
  fi
  if [[ -f "${ENV_FILE}" ]]; then
    if [[ -n "${DEPLOY_SERVICE}" ]]; then
      docker compose --env-file "${ENV_FILE}" up -d --build "${DEPLOY_SERVICE}"
    else
      docker compose --env-file "${ENV_FILE}" up -d --build
    fi
  else
    if [[ -n "${DEPLOY_SERVICE}" ]]; then
      docker compose up -d --build "${DEPLOY_SERVICE}"
    else
      docker compose up -d --build
    fi
  fi
elif [[ "${deployment_mode}" == "dockerfile" ]]; then
  docker build -t "steellink-admin:${TARGET_ENV}" .
fi

echo "deployment command executed"
