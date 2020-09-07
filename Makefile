# fdwctl Makefile
APPVERSION=0.0.1

build: lint test
	CGO_ENABLED=0 go build -i -pkgdir "$(shell go env GOPATH)/pkg" -installsuffix nocgo -ldflags "-s -w -X main.cmd.AppVersion=$(APPVERSION)" -o fdwctl ./cmd/fdwctl
	{ [ "$(shell uname -s)" != "Darwin" ] && which upx >/dev/null 2>&1 && upx -q fdwctl; } || true

.PHONY: build build-docker clean start-docker stop-docker restart-docker lint test

build-docker: clean lint test
	docker build --no-cache --build-arg "APPVERSION=$(APPVERSION)" -t "neflyte/fdwctl:$(APPVERSION)" .

clean:
	{ [ -f "./fdwctl" ] && rm -f ./fdwctl; } || true

start-docker:
	docker-compose -f testdata/docker-compose.yaml up -d

stop-docker:
	docker-compose -f testdata/docker-compose.yaml down -v

restart-docker: stop-docker start-docker
	@echo "services restarted."

lint:
	golangci-lint run

test:
	CGO_ENABLED=0 go test ./...
