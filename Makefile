# fdwctl Makefile
APPVERSION=0.0.4

.PHONY: build build-docker clean start-docker stop-docker restart-docker lint test install test-dstate-yaml test-dstate-json reformat reformat-gofmt reformat-goimports

build:
	CGO_ENABLED=0 go build -ldflags "-s -w -X github.com/neflyte/fdwctl/cmd/fdwctl/cmd.AppVersion=$(APPVERSION)" -o fdwctl ./cmd/fdwctl
	@hash upx 2>/dev/null && { upx -q fdwctl || true; }

build-docker: build
	docker buildx build --build-arg "APPVERSION=$(APPVERSION)" -t "neflyte/fdwctl:$(APPVERSION)" -t "neflyte/fdwctl:latest" .

clean:
	{ [ -f ./fdwctl ] && rm -f ./fdwctl; } || true

start-docker:
	docker-compose -f testdata/docker-compose.yaml up -d

stop-docker:
	docker-compose -f testdata/docker-compose.yaml down -v

restart-docker: stop-docker start-docker
	@echo "services restarted."

lint: check-fieldalignment
	@golangci-lint --version
	golangci-lint run --timeout=10m --verbose

test:
	@mkdir -p coverage
	@{ [ -r coverage/c.out ] && rm -f coverage/c.out; } || true
	CGO_ENABLED=0 go test -covermode=count -coverprofile=coverage/c.out ./...
	@{ [ -r coverage/coverage-heatmap.html ] && rm -f coverage/coverage-heatmap.html; } || true
	go tool cover -html=coverage/c.out -o coverage/coverage-heatmap.html

install: clean build
	cp ./fdwctl "$(shell go env GOPATH)/bin"

test-dstate-yaml:
	./fdwctl --config testdata/dstate.yaml apply
	hash psql 2>/dev/null && { \
  		PGPASSWORD='passw0rd' psql -h localhost -p 5432 -U fdw -d fdw -c 'SELECT * FROM remotedb.foo;'; \
	}

test-dstate-json:
	./fdwctl --config testdata/dstate.json apply
	hash psql 2>/dev/null && { \
  		PGPASSWORD='passw0rd' psql -h localhost -p 5432 -U fdw -d fdw -c 'SELECT * FROM remotedb.foo;'; \
	}

reformat: reformat-gofmt reformat-goimports
	@echo "reformatted source files."

reformat-gofmt:
	go fmt ./...

reformat-goimports:
	@hash goimports 2>/dev/null || { cd && go install golang.org/x/tools/cmd/goimports@latest; cd -; }
	find . -type f -name "*.go" | xargs goimports -w

outdated:
	hash go-mod-outdated 2>/dev/null || { cd && go install github.com/psampaz/go-mod-outdated@v0.9.0; }
	go list -json -u -m all | go-mod-outdated -direct -update

ensure-fieldalignment:
	hash fieldalignment 2>/dev/null || { cd && go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest; }

check-fieldalignment: ensure-fieldalignment
	fieldalignment ./...

autofix-fieldalignment: ensure-fieldalignment
	fieldalignment -fix ./...
