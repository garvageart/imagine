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


# Connect using the configured initial superuser (POSTGRES_USER). The official
# image creates the initial superuser with the name in POSTGRES_USER, so
# attempting to connect as literal "postgres" may fail when a different user
# is supplied via environment variables.
psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname "postgres" <<-SQL
DO
\$\$
BEGIN
  -- Create or alter role matching POSTGRES_USER
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '${POSTGRES_USER}') THEN
    EXECUTE format('CREATE ROLE %I WITH LOGIN SUPERUSER PASSWORD %L', '${POSTGRES_USER}', '${POSTGRES_PASSWORD}');
  ELSE
    EXECUTE format('ALTER ROLE %I WITH SUPERUSER PASSWORD %L', '${POSTGRES_USER}', '${POSTGRES_PASSWORD}');
  END IF;

  -- Create database if it doesn't exist, and set owner
END
\$\$;
SQL

# The CREATE DATABASE command cannot be executed inside a PL/pgSQL function/DO
# block. Run database creation as separate commands so they are executed as
# top-level SQL statements. Keep role creation/alteration in the DO block
# above (it is allowed) and create databases idempotently from shell.

# Helper to run a query and return non-empty on match
exists_db() {
  psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname "postgres" -At -c "SELECT 1 FROM pg_database WHERE datname='${1}'" | grep -q 1
}

# Create `${POSTGRES_DB}` if it doesn't exist
if ! exists_db "${POSTGRES_DB}"; then
  psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname "postgres" -c "CREATE DATABASE \"${POSTGRES_DB}\" OWNER \"${POSTGRES_USER}\";"
fi

# Also ensure a database named after the user exists
if ! exists_db "${POSTGRES_USER}"; then
  psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname "postgres" -c "CREATE DATABASE \"${POSTGRES_USER}\" OWNER \"${POSTGRES_USER}\";"
fi

exit 0
