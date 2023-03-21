.PHONY: test clean

build:
	go build -o bin/main main.go

run:
	bin/main

go-run:
	go run main.go

deps:
	go get

tidy:
	go mod tidy

clean: tidy
	rm -rf bin

test:
	go test -v ./test/...

dist:
	echo "Compiling for other platforms"
	GOOS=linux GOARCH=arm go build -o bin/main-linux-arm main.go
	GOOS=linux GOARCH=arm64 go build -o bin/main-linux-arm64 main.go
	GOOS=freebsd GOARCH=386 go build -o bin/main-freebsd-386 main.go
	GOOS=windows GOARCH=arm64 go build -o bin/main-windows-arm64 main.go

all: clean deps build test dist run