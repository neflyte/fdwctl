#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    \c remotedb
    CREATE TYPE public.enum_type AS ENUM ('enum_one', 'enum_two', 'enum_three');
    CREATE TABLE public.foo (id int, name text, enum_value public.enum_type);
EOSQL
