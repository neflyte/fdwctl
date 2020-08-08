# fdwctl
A PostgreSQL Foreign Data Wrapper (FDW) management CLI


### What is it?

`fdwctl` is a command line tool to manage foreign servers, user mappings, and importing remote schemas using the PostgreSQL Foreign Data Wrapper (FDW).

### What does it do?

`fdwctl` allows create, read, update, and delete (CRUD) operations on foreign servers, user mappings, and importing remote schemas with reasonably simple commands. It also can attempt to create local ENUM types that are used by a remote schema before importing that schema.

The CLI interface is written with the original intention of being used as part of Kubernetes deployments.

### Features

- Supports PostgreSQL `postgres_fdw` wrapper; support for other wrappers is pending
- Can apply a desired configuration state to a database
- Single, statically-linked binary; can be compiled on any platform supporting Golang 1.14+
- Multiple log message formats (Text, JSON; Elasticstack schema support is pending)

### Usage

```
Usage:
  fdwctl [command]

Available Commands:
  apply       Apply a desired state
  create      Create objects
  drop        Drop (delete) objects
  edit        Edit objects
  help        Help about any command
  list        List objects

Flags:
      --config string       location of program configuration file
      --connection string   database connection string
  -h, --help                help for fdwctl
      --logformat string    log output format [text, json] (default "text")
      --loglevel string     log message level [trace, debug, info, warn, error, fatal, panic] (default "trace")
      --nologo              suppress program name and version message

Use "fdwctl [command] --help" for more information about a command.
```

### Example Configuration

The following is an example of the application configuration file in YAML format. The equivalent in JSON format is also supported.

```yaml
FDWConnection: "host=localhost port=5432 dbname=fdw user=fdw sslmode=disable"
FDWConnectionSecret:
  value: "passw0rd"
DesiredState:
  Extensions:
    - name: postgres_fdw
  Servers:
    - name: remotedb
      host: remotedb1
      port: 5432
      db: remotedb
      UserMap:
        - localuser: fdw
          remoteuser: remoteuser
          remotesecret:
            value: "r3m0TE!"
            #fromEnv: "REMOTEDB_CREDENTIAL"
            #fromFile: /tmp/credential.txt
            #fromK8s:
              #namespace: default
              #secretName: my-secret-object
              #secretKey: postgresql-password
      Schemas:
        - localschema: remotedb
          remoteschema: public
          importenums: true
          enumconnection: "postgres://remoteuser@localhost:15432/remotedb?sslmode=disable"
          enumsecret:
            value: "r3m0TE!"
            #fromEnv: "REMOTEDB_CREDENTIAL"
            #fromFile: /tmp/credential.txt
            #fromK8s:
              #namespace: default
              #secretName: my-secret-object
              #secretKey: postgresql-password
```
