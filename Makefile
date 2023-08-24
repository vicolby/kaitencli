build:
	@go build -o bin/kaitencli

run: build
	@./bin/kaitencli 

test:
	@go test -v ./...

