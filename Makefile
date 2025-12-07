.PHONY: build build-fetch build-ics run fetch ics clean

# Build all binaries
build: build-fetch build-ics

build-fetch:
	@go build -o bin/mma-fetch ./cmd/fetch

build-ics:
	@go build -o bin/mma-ics ./cmd/ics

# Run commands
run: fetch ics

fetch:
	@go run ./cmd/fetch

ics:
	@go run ./cmd/ics

clean:
	@rm -rf bin/ events.json ufc-events.ics
