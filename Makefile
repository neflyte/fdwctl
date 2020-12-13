# fdwctl Makefile
APPVERSION=0.0.3

.PHONY: build build-docker clean start-docker stop-docker restart-docker lint test install

build: lint test
	CGO_ENABLED=0 go build -i -pkgdir "$(shell go env GOPATH)/pkg" -installsuffix nocgo -ldflags "-s -w -X main.cmd.AppVersion=$(APPVERSION)" -o fdwctl ./cmd/fdwctl
	type -p upx >/dev/null 2>&1 && upx -q fdwctl

build-docker: clean lint test
	docker build --no-cache --build-arg "APPVERSION=$(APPVERSION)" -t "neflyte/fdwctl:$(APPVERSION)" -t "neflyte/fdwctl:latest" .

clean:
	{ [ -f ./fdwctl ] && rm -f ./fdwctl; } || true

start-docker:
	docker-compose -f testdata/docker-compose.yaml up -d

stop-docker:
	docker-compose -f testdata/docker-compose.yaml down -v

restart-docker: stop-docker start-docker
	@echo "services restarted."

lint:
	golangci-lint run

test:
	{ [ -r coverage/c.out ] && rm -f coverage/c.out; } || true
	CGO_ENABLED=0 go test -covermode=count -coverprofile=coverage/c.out ./...
	{ [ -r coverage/coverage-heatmap.html ] && rm -f coverage/coverage-heatmap.html; } || true
	go tool cover -html=coverage/c.out -o coverage/coverage-heatmap.html

install: clean build
	cp ./fdwctl "$(shell go env GOPATH)/bin"
