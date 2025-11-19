#!/bin/bash
set -euo pipefail

# This script runs during Postgres container initialization (first time only).
# It ensures a role exists matching POSTGRES_USER and grants SUPERUSER privileges.
# Uses the environment variables provided by the postgres image: POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB

# If any variable is empty, exit early
: "${POSTGRES_USER:?POSTGRES_USER is not set}"
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is not set}"
: "${POSTGRES_DB:?POSTGRES_DB is not set}"

# Wait a moment if postgres isn't ready (the official image runs these scripts after initdb,
# but we add a small retry to be safe when executed interactively).

for i in 1 2 3 4 5; do
  if pg_isready -q -d "postgres"; then
    break
  fi
  sleep 1
done

# Create role if it doesn't exist, and set superuser and password
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-SQL || true
DO
\$do\$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '$POSTGRES_USER') THEN
    CREATE ROLE "$POSTGRES_USER" WITH LOGIN SUPERUSER PASSWORD '$POSTGRES_PASSWORD';
  ELSE
    -- If role exists, alter it to ensure superuser and password
    ALTER ROLE "$POSTGRES_USER" WITH SUPERUSER PASSWORD '$POSTGRES_PASSWORD';
  END IF;
END
\$do\$;

-- Ensure the database exists and is owned by the role
CREATE DATABASE "$POSTGRES_DB" OWNER "$POSTGRES_USER";

SQL

exit 0
