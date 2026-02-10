SHELL := /bin/bash
.PHONY: build build-fetch build-ics build-web run fetch ics web clean docker

build: build-fetch build-ics build-web

build-fetch:
	@go build -o bin/ufc-fetch ./cmd/fetch

build-ics:
	@go build -o bin/ufc-ics ./cmd/ics

build-web:
	@go build -o bin/ufc-web ./cmd/web

run: fetch ics

fetch:
	@mkdir -p assets
	@go run ./cmd/fetch

ics:
	@go run ./cmd/ics

web:
	@go run ./cmd/web

clean:
	@rm -rf bin/ assets/

docker:
	@docker build -t ufc .
