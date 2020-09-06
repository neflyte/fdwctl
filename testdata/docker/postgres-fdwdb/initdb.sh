#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
   \c fdw
   CREATE USER fdw WITH PASSWORD 'fDw!u5eR';
   GRANT ALL PRIVILEGES ON DATABASE fdw TO fdw;
EOSQL
