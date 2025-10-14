APP        := at-ping
BIN_LOCAL  := bin/$(APP)
MAIN       := ./cmd
GO         ?= go

.PHONY: all build run test clean install uninstall restart logs

all: build

build:
	@mkdir -p bin
	$(GO) build -o $(BIN_LOCAL) $(MAIN)

run: build
	$(BIN_LOCAL) $(ARGS)

test:
	$(GO) test ./...

clean:
	rm -rf bin

install: build
	sudo bash service/install.sh $(BIN_LOCAL)

uninstall:
	sudo bash service/uninstall.sh

restart:
	sudo systemctl restart $(APP)

logs:
	sudo journalctl -u $(APP) -f
