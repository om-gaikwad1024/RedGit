.PHONY: build test install clean run

build:
	go build -o bin/redgit cmd/redgit/main.go

test:
	go test ./...

install:
	go install cmd/redgit/main.go

clean:
	rm -rf bin/
	rm -rf test-repo/

run:
	go run cmd/redgit/main.go

fmt:
	go fmt ./...