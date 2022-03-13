#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
   \c fdw
   CREATE USER fdw WITH SUPERUSER PASSWORD 'passw0rd';
   ALTER DATABASE fdw OWNER TO fdw;
   GRANT ALL PRIVILEGES ON DATABASE fdw TO fdw;
EOSQL
