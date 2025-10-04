build:
	@go build -o mma-cal .

dev: build
	@go run .

