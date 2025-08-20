.PHONY: api-generate

api-generate:
	go mod tidy
	go mod download
	go generate ./...