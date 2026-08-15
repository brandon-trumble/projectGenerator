BINARY_NAME=projectGeneratorExe

build:
	go build -o $(BINARY_NAME)

# copies the binary to a user-local bin dir and puts that dir on PATH
install: build
	./$(BINARY_NAME) --install

clean:
	rm $(BINARY_NAME)

run:
	go run main.go

vet:
	go vet