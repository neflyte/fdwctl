#!/bin/bash
set -e
# Create the remote user and grant privileges
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    \c remotedb
    CREATE USER remoteuser WITH PASSWORD 'r3m0TE!';
    GRANT ALL PRIVILEGES ON SCHEMA public TO remoteuser;
EOSQL
# Seed the remotedb database with some test data
psql -v ON_ERROR_STOP=1 --username "remoteuser" --dbname "remotedb" <<-EOSQL
    \c remotedb
    CREATE TYPE public.enum_type AS ENUM ('enum_one', 'enum_two', 'enum_three');
    CREATE TABLE public.foo (id int, name text, enum_value public.enum_type);
    INSERT INTO public.foo (id, name, enum_value) VALUES (1, 'FOO', 'enum_two');
EOSQL
