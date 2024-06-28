export PATH := $(PATH):`go env GOPATH`/bin
export GO111MODULE=on
LDFLAGS := -s -w

all: env fmt build

build: tuns tunc

env:
	@go version

fmt:
	@go fmt ./...

tuns:
	env CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -tags tuns -o bin/tuns ./cmd/tuns

tunc:
	env CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -tags tunc -o bin/tunc ./cmd/tunc

clean:
	rm -f ./bin/tunc
	rm -f ./bin/tuns
	rm -rf ./lastversion