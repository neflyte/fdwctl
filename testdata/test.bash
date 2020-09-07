#!/usr/bin/env bash
trap exit ERR
set -x
../fdwctl --config config.yaml create extension postgres_fdw
../fdwctl --config config.yaml create server remotedb1 --serverhost remotedb1 --serverport 5432 --serverdbname remotedb
../fdwctl --config config.yaml create usermap --servername remotedb1 --localuser postgres --remoteuser remoteuser --remotepassword "r3m0TE!"
../fdwctl --config config.yaml create usermap --servername remotedb1 --localuser fdw --remoteuser remoteuser --remotepassword "r3m0TE!"
../fdwctl --config config.yaml create schema --servername remotedb1 --localschema remotedb1 --remoteschema public --importenums --enumconnection "postgres://postgres:passw0rd@localhost:15432/remotedb?sslmode=disable"
type -p psql &>/dev/null && {
  PGPASSWORD='passw0rd' psql -h localhost -p 5432 -U postgres -d fdw -c 'GRANT USAGE ON SCHEMA remotedb1 TO postgres;'
  PGPASSWORD='passw0rd' psql -h localhost -p 5432 -U postgres -d fdw -c 'GRANT USAGE ON SCHEMA remotedb1 TO fdw;'
  PGPASSWORD='passw0rd' psql -h localhost -p 5432 -U postgres -d fdw -c 'GRANT SELECT ON ALL TABLES IN SCHEMA remotedb1 TO postgres;'
  PGPASSWORD='passw0rd' psql -h localhost -p 5432 -U postgres -d fdw -c 'GRANT SELECT ON ALL TABLES IN SCHEMA remotedb1 TO fdw;'
  PGPASSWORD='fDw!u5eR' psql -h localhost -p 5432 -U fdw -d fdw -c 'SELECT * FROM remotedb1.foo;'
}
echo "done."
