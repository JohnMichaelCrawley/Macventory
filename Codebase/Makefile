.PHONY: build test fmt vet check
build:
	go build -trimpath - ldflags="-s -w" -o bin/macventory .
test:
	go test ./...
fmt:
	gofmt -w . 
vet: 
	go vet ./...
check fmt vet test