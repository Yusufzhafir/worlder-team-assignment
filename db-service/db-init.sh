#!/usr/bin/env bash
set -euo pipefail

# Defaults match docker-compose.yml; override via env if needed.
MYSQL_CONTAINER=${MYSQL_CONTAINER:-db-service}
MYSQL_DB=${MYSQL_DATABASE:-sensors}
MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD:-root}

# Ensure DB is up
echo "Waiting for MySQL to be healthy..."
until docker compose ps --services --filter "status=running" | grep -q "^${MYSQL_CONTAINER}\$"; do
  sleep 1
done

echo "Applying schema from db-service/schema.sql to ${MYSQL_DB}..."
docker compose exec -T "${MYSQL_CONTAINER}" \
  sh -c "mysql -uroot -p\"${MYSQL_ROOT_PASSWORD}\" ${MYSQL_DB}" \
  < db-service/schema.sql

echo "Done."
