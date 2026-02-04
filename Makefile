BINARY_NAME=burrow

all: build

build:
	go build -o $(BINARY_NAME) ./cmd/burrow/main.go

clean:
	rm -f $(BINARY_NAME)

install: build
	mv $(BINARY_NAME) /usr/local/bin/

test:
	go test ./...

.PHONY: all build clean install test
