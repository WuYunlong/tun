export PATH := $(PATH):`go env GOPATH`/bin
export GO111MODULE=on
LDFLAGS := -s -w

all: env fmt build

env:
	@go version

fmt:
	go fmt ./...

build: tuns tunc

tuns:
	env CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -tags tuns -o bin/tuns ./cmd/tuns

tunc:
	env CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -tags tunc -o bin/tunc ./cmd/tunc

clean:
	rm -f ./bin/tuns
	rm -f ./bin/tunc