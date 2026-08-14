BINARY_NAME=projectGeneratorExe

build:
	go build -o $(BINARY_NAME)

clean:
	rm $(BINARY_NAME)

run:
	go run main.go

vet:
	go vet