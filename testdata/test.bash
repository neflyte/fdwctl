#!/usr/bin/env bash
test -f ./fdwctl || {
  echo "*  build fdwctl before running this test"
  exit 1
}
trap exit ERR
CONFIGFILE="testdata/config.yaml"
set -x
./fdwctl --config ${CONFIGFILE} create extension postgres_fdw
./fdwctl --config ${CONFIGFILE} create server remotedb1 --serverhost remotedb1 --serverport 5432 --serverdbname remotedb
./fdwctl --config ${CONFIGFILE} create usermap --servername remotedb1 --localuser postgres --remoteuser remoteuser --remotepassword "r3m0TE!"
./fdwctl --config ${CONFIGFILE} create usermap --servername remotedb1 --localuser fdw --remoteuser remoteuser --remotepassword "r3m0TE!"
./fdwctl --config ${CONFIGFILE} create schema --servername remotedb1 --localschema remotedb1 --remoteschema public --importenums --enumconnection "postgres://remoteuser:r3m0TE!@localhost:15432/remotedb?sslmode=disable"
hash psql 2>/dev/null && {
  PGPASSWORD='passw0rd' psql -h localhost -p 5432 -U postgres -d fdw -c 'GRANT USAGE ON FOREIGN SERVER remotedb1 TO postgres;'
  PGPASSWORD='passw0rd' psql -h localhost -p 5432 -U postgres -d fdw -c 'GRANT USAGE ON FOREIGN SERVER remotedb1 TO fdw;'
  PGPASSWORD='passw0rd' psql -h localhost -p 5432 -U postgres -d fdw -c 'GRANT USAGE ON SCHEMA remotedb1 TO postgres;'
  PGPASSWORD='passw0rd' psql -h localhost -p 5432 -U postgres -d fdw -c 'GRANT USAGE ON SCHEMA remotedb1 TO fdw;'
  PGPASSWORD='passw0rd' psql -h localhost -p 5432 -U postgres -d fdw -c 'GRANT SELECT ON ALL TABLES IN SCHEMA remotedb1 TO postgres;'
  PGPASSWORD='passw0rd' psql -h localhost -p 5432 -U postgres -d fdw -c 'GRANT SELECT ON ALL TABLES IN SCHEMA remotedb1 TO fdw;'
  PGPASSWORD='passw0rd' psql -h localhost -p 5432 -U fdw -d fdw -c 'SELECT * FROM remotedb1.foo;'
}
echo "done."
