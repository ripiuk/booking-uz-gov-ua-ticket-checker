.PHONY: update build build-windows test test-cov

update:
	@echo "Updating dependencies"
	@cat requirements.txt | xargs -l1 go get

build:
	@echo "Create build for Linux"
	mkdir -p build
	GOOS=linux GOARCH=amd64 go build -o build/uzTicketsMonitoring main.go

build-windows:
	@echo "Create build for Windows"
	mkdir -p build
	GOOS=windows GOARCH=amd64 go build -o build/uzTicketsMonitoring.exe main.go

test:
	@echo "Run tests"
	@go test

test-cov:
	@echo "Run tests with coverage"
	@go test -coverprofile=/tmp/cover.out && go tool cover -html=/tmp/cover.out

clean:
	@echo "Cleanup"
	@rm -fr /tmp/cover*