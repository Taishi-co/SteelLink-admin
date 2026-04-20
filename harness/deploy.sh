#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TARGET_ENV="${1:-}"
ENV_FILE="${ENV_FILE:-.env.${TARGET_ENV}}"
DEPLOY_SERVICE="${DEPLOY_SERVICE:-}"
STATIC_OUT_DIR="${STATIC_OUT_DIR:-dist}"

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

planned_command=""

if [[ -n "${DEPLOY_COMMAND:-}" ]]; then
  planned_command="${DEPLOY_COMMAND}"
  echo "deployment entrypoint: DEPLOY_COMMAND"
elif [[ -n "${compose_file}" && -f "Dockerfile" ]]; then
  if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    if [[ -f "${ENV_FILE}" ]]; then
      docker compose --env-file "${ENV_FILE}" config >/dev/null
      echo "validated compose plan with ${ENV_FILE}"
      if [[ -n "${DEPLOY_SERVICE}" ]]; then
        planned_command="docker compose --env-file ${ENV_FILE} up -d --build ${DEPLOY_SERVICE}"
      else
        planned_command="docker compose --env-file ${ENV_FILE} up -d --build"
      fi
    else
      docker compose config >/dev/null
      echo "validated compose plan with default environment"
      if [[ -n "${DEPLOY_SERVICE}" ]]; then
        planned_command="docker compose up -d --build ${DEPLOY_SERVICE}"
      else
        planned_command="docker compose up -d --build"
      fi
    fi
  else
    echo "docker compose file exists but docker compose is not available locally"
    if [[ -n "${DEPLOY_SERVICE}" ]]; then
      planned_command="docker compose up -d --build ${DEPLOY_SERVICE}"
    else
      planned_command="docker compose up -d --build"
    fi
  fi
elif [[ -f "Dockerfile" ]]; then
  planned_command="docker build -t steellink-admin:${TARGET_ENV} ."
else
  echo "no deploy entrypoint detected; set DEPLOY_COMMAND or add Dockerfile/Compose config"
fi

if [[ -n "${planned_command}" ]]; then
  echo "planned command: ${planned_command}"
fi

if [[ "${DEPLOY_EXECUTE:-0}" != "1" ]]; then
  echo "dry-run mode"
  echo "set DEPLOY_EXECUTE=1 to enable real deployment steps"
  exit 0
fi

if [[ -z "${planned_command}" ]]; then
  echo "deployment command is not configured" >&2
  exit 1
fi

if [[ -n "${DEPLOY_COMMAND:-}" ]]; then
  bash -lc "${DEPLOY_COMMAND}"
elif [[ -n "${compose_file}" && -f "Dockerfile" ]]; then
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
elif [[ -f "Dockerfile" ]]; then
  docker build -t "steellink-admin:${TARGET_ENV}" .
fi

echo "deployment command executed"
