#!/usr/bin/env bash
# for local development

POSTGRES_IMAGE_TAG="${PERCONA_IMAGE_TAG:-18-alpine3.23}"
POSTGRES_ROOT_PASSWORD="${PERCONA_ROOT_PASSWORD:-postgres}"

set -eu

if ! podman volume exists "postgres-data"; then
  podman volume create "postgres-data"
fi

exec podman run -d --rm --name "local-postgres" -p "5432:5432" \
    -v "postgres-data:/var/lib/postgresql:z" \
    -e POSTGRES_PASSWORD="${POSTGRES_ROOT_PASSWORD}" \
    "docker.io/library/postgres:${POSTGRES_IMAGE_TAG}"