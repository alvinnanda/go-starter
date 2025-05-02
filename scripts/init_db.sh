#!/bin/bash

# Database credentials
DB_USER=${POSTGRES_USER:-postgres}
DB_PASSWORD=${POSTGRES_PASSWORD:-1sampai6}
DB_NAME=${POSTGRES_DB:-starter_db}
DB_PORT=${POSTGRES_PORT:-5432}
DB_HOST=${POSTGRES_HOST:-localhost}

# Check if psql is installed
if ! [ -x "$(command -v psql)" ]; then
  echo 'Error: psql is not installed.' >&2
  exit 1
fi

# Set PGPASSWORD environment variable for authentication
export PGPASSWORD=$DB_PASSWORD

# Test connection to PostgreSQL server
echo "Testing connection to PostgreSQL server..."
if ! psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -c '\q' 2>/dev/null; then
  echo "Error: Could not connect to PostgreSQL server. Check your credentials and make sure the server is running." >&2
  exit 1
fi

# Check if the database exists
if ! psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -lqt | cut -d \| -f 1 | grep -qw "$DB_NAME"; then
  echo "Creating database: $DB_NAME"
  # Create database
  psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -c "CREATE DATABASE $DB_NAME;"
  if [ $? -ne 0 ]; then
    echo "Error: Failed to create database." >&2
    exit 1
  fi
else
  echo "Database $DB_NAME already exists."
fi

# Run the migration script
echo "Running the migration script"
psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d "$DB_NAME" -f ./scripts/migrations/init.sql
if [ $? -ne 0 ]; then
  echo "Error: Failed to run migration script." >&2
  exit 1
fi

echo "Database initialization completed successfully"
