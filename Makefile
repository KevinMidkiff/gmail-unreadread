.PHONY: build run clean

all: build

build:
	go build -o bin/ .

run:
	go run .

clean:
	go clean
