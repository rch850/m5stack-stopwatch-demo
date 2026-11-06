TINYGO ?= tinygo
GO ?= go
TARGET := ./targets/stopwatch.json

.PHONY: build flash test preview

build:
	mkdir -p build
	$(TINYGO) build -target=$(TARGET) -o build/crab.bin .

# This overwrites the installed firmware. Set PORT=/dev/cu.usbmodem... .
flash:
	$(TINYGO) flash -target=$(TARGET) $(if $(PORT),-port=$(PORT)) .

test:
	$(GO) test ./...

preview:
	$(GO) run ./cmd/preview
