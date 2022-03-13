# fdwctl Makefile
APPVERSION=0.0.4

.PHONY: build build-docker clean start-docker stop-docker restart-docker lint test install test-dstate-yaml

build:
	CGO_ENABLED=0 go build -ldflags "-s -w -X github.com/neflyte/fdwctl/cmd/fdwctl/cmd.AppVersion=$(APPVERSION)" -o fdwctl ./cmd/fdwctl
	@hash upx 2>/dev/null && { upx -q fdwctl || true; }

build-docker: build
	docker build --build-arg "APPVERSION=$(APPVERSION)" -t "neflyte/fdwctl:$(APPVERSION)" -t "neflyte/fdwctl:latest" .

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
	@mkdir -p coverage
	@{ [ -r coverage/c.out ] && rm -f coverage/c.out; } || true
	CGO_ENABLED=0 go test -covermode=count -coverprofile=coverage/c.out ./...
	@{ [ -r coverage/coverage-heatmap.html ] && rm -f coverage/coverage-heatmap.html; } || true
	go tool cover -html=coverage/c.out -o coverage/coverage-heatmap.html

install: clean build
	cp ./fdwctl "$(shell go env GOPATH)/bin"

test-dstate-yaml:
	./fdwctl --config testdata/dstate.yaml apply
