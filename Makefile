# fdwctl Makefile
APPVERSION=0.0.1

build: lint
	CGO_ENABLED=0 go build -i -pkgdir "$(GOPATH)/pkg" -installsuffix nocgo -ldflags "-s -w -X main.cmd.AppVersion=$(APPVERSION)" -o fdwctl ./cmd/fdwctl
	type -p upx >/dev/null && upx -q fdwctl

build-docker: clean lint
	docker build --no-cache -t "neflyte/fdwctl:$(APPVERSION)" .

clean:
	[ -f "./fdwctl" ] && rm -f ./fdwctl

start-docker:
	docker-compose -f testdata/docker-compose.yaml up -d

stop-docker:
	docker-compose -f testdata/docker-compose.yaml down -v

lint:
	golangci-lint run -E gosec
