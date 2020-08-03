FROM golang:1.14-buster AS builder
RUN apt-get update && apt-get install --yes upx-ucl
WORKDIR /go/src/fdwctl
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags "-s -w" -o fdwctl ./cmd/fdwctl
RUN upx -q fdwctl

FROM scratch
COPY --from=builder /go/src/fdwctl/fdwctl /usr/local/bin
ENTRYPOINT ["/usr/local/bin/fdwctl"]
