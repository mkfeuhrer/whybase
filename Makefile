.PHONY: test build install vet

test:
	go test ./...

vet:
	go vet ./...

build: vet
	go build -o whybase ./cmd/whybase

install:
	go install ./cmd/whybase
