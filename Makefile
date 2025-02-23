.PHONY: build run clean

all: build

build:
	go build -o bin/ ./cmd/unreadread/

run:
	go run ./cmd/unreadread

clean:
	go clean
	rm -r ./bin/
	rm token.json
