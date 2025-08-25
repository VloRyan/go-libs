test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run

fmt:
	gofumpt -l -w .

fmt-check:
	gofumpt -d .
