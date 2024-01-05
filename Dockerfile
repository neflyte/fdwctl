ARG APPVERSION

FROM golang:1.18 AS builder
RUN apt-get update --yes && apt-get install --yes upx-ucl
WORKDIR /tmp/src/fdwctl
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X github.com/neflyte/fdwctl/cmd/fdwctl/cmd.AppVersion=${APPVERSION}" -o fdwctl ./cmd/fdwctl
RUN upx -q fdwctl

FROM scratch
COPY --from=builder /tmp/src/fdwctl/fdwctl /usr/local/bin/
ENTRYPOINT [ "/usr/local/bin/fdwctl" ]
