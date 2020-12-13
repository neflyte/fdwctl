ARG APPVERSION

FROM golang:1.12-buster AS builder
RUN apt-get update && apt-get install --yes upx-ucl
WORKDIR /go/src/fdwctl
COPY . .
RUN GO111MODULE=on go mod download
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags "-s -w -X main.cmd.AppVersion=${APPVERSION}" -o fdwctl ./cmd/fdwctl
RUN upx -q fdwctl

FROM scratch
COPY --from=builder /go/src/fdwctl/fdwctl /usr/local/bin/
ENTRYPOINT [ "/usr/local/bin/fdwctl" ]
