.PHONY: build clean deploy

BINARY_NAME=scaling-parakeet

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/${BINARY_NAME} main.go

clean:
	go clean

run: build
	bin/${BINARY_NAME}

test:
	go test ./...

test-coverage:
	go test ./... -coverprofile=coverage.out
