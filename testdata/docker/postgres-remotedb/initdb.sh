#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    \c remotedb
    CREATE USER remoteuser WITH PASSWORD 'r3m0TE!';
    ALTER DATABASE remotedb OWNER TO remoteuser;
    GRANT ALL PRIVILEGES ON DATABASE remotedb TO remoteuser;
    CREATE TYPE public.enum_type AS ENUM ('enum_one', 'enum_two', 'enum_three');
    ALTER TYPE public.enum_type OWNER TO remoteuser;
    CREATE TABLE public.foo (id int, name text, enum_value public.enum_type);
    ALTER TABLE public.foo OWNER TO remoteuser;
    CREATE SCHEMA "miXEDcaSEscHEMa";
    ALTER SCHEMA "miXEDcaSEscHEMa" OWNER TO remoteuser;
    CREATE TYPE "miXEDcaSEscHEMa"."enum_MixedCase" AS ENUM ('ENUm_ONE', 'enuM-2', 'enumTHREE');
    ALTER TYPE "miXEDcaSEscHEMa"."enum_MixedCase" OWNER TO remoteuser;
    CREATE TABLE "miXEDcaSEscHEMa"."Table-MixedCase" (id int, name text, "mixedCASEenum" "miXEDcaSEscHEMa"."enum_MixedCase");
    ALTER TABLE "miXEDcaSEscHEMa"."Table-MixedCase" OWNER TO remoteuser;
EOSQL
