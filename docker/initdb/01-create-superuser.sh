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


# Run psql as the default superuser 'postgres' (guaranteed to exist during init)
psql -v ON_ERROR_STOP=1 --username "postgres" <<-'SQL'
DO
$$
BEGIN
  -- Create or alter role
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '${POSTGRES_USER}') THEN
    EXECUTE format('CREATE ROLE %I WITH LOGIN SUPERUSER PASSWORD %L', '${POSTGRES_USER}', '${POSTGRES_PASSWORD}');
  ELSE
    EXECUTE format('ALTER ROLE %I WITH SUPERUSER PASSWORD %L', '${POSTGRES_USER}', '${POSTGRES_PASSWORD}');
  END IF;
  -- Create database if it doesn't exist, and set owner
  IF NOT EXISTS (SELECT FROM pg_database WHERE datname = '${POSTGRES_DB}') THEN
    EXECUTE format('CREATE DATABASE %I OWNER %I', '${POSTGRES_DB}', '${POSTGRES_USER}');
  END IF;
END
$$;
SQL

exit 0
