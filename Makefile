BINARY_NAME=guh

.PHONY: build clean test

build:
	go build -o $(BINARY_NAME) .

clean:
	rm -f $(BINARY_NAME)

test:
	go test ./...
