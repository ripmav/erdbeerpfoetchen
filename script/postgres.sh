#!/usr/bin/env bash
# for local development
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

POSTGRES_IMAGE_NAME="${POSTGRES_IMAGE_NAME:-library/postgres}"
POSTGRES_IMAGE_TAG="${POSTGRES_IMAGE_TAG:-18-alpine3.23}"
POSTGRES_ROOT_PASSWORD="${PERCONA_ROOT_PASSWORD:-postgres}"

set -eu

if ! podman volume exists "postgres-data"; then
  podman volume create "postgres-data"
fi

exec podman run -d --rm --name "local-postgres" -p "5432:5432" \
    -v "postgres-data:/var/lib/postgresql:z" \
    -v "${SCRIPT_DIR}/config/postgresql.conf:/etc/postgresql/postgresql.conf:z" \
    -v "${SCRIPT_DIR}/config/pg_hba.conf:/etc/postgresql/pg_hba.conf:z" \
    -e POSTGRES_PASSWORD="${POSTGRES_ROOT_PASSWORD}" \
    "docker.io/${POSTGRES_IMAGE_NAME}:${POSTGRES_IMAGE_TAG}" \
    -c config_file=/etc/postgresql/postgresql.conf \
    -c hba_file=/etc/postgresql/pg_hba.conf