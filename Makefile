APP := at-ping
BIN := bin/$(APP)

.PHONY: all build run test clean

all: build

build:
	mkdir -p bin
	go build -o $(BIN) ./cmd

run: build
	$(BIN) $(ARGS)

test:
	go test ./...

clean:
	rm -rf bin
