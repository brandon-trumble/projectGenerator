BINARY_NAME=projectGeneratorExe

build:
	go build -o $(BINARY_NAME)

# copies the binary to a user-local bin dir and puts that dir on PATH
install: build
	./$(BINARY_NAME) --install

# validates .goreleaser.yaml without building anything
release-check:
	goreleaser check

# builds the release archives for every platform into dist/, publishing nothing
snapshot:
	goreleaser release --snapshot --clean

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist

run:
	go run main.go

vet:
	go vet ./...

test:
	go test ./...


