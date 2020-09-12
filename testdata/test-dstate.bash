#!/usr/bin/env bash
test -f ./fdwctl || {
  echo "*  build fdwctl before running this test"
  exit 1
}
trap exit ERR
set -x
./fdwctl --config testdata/dstate.yaml apply
type -p psql &>/dev/null && {
  PGPASSWORD='fDw!u5eR' psql -h localhost -p 5432 -U fdw -d fdw -c 'SELECT * FROM remotedb.foo;'
}
echo "done."
