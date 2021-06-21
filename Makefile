BIN_NAME=./bin/lignum
ALL_GO_FILES=$(shell find . -type f  -name '*.go')
CONFIG_FILE="config.yml"
export ENV=development

.PHONY: test

clean:
	@test ! -e bin || rm -r bin

build-linux:clean
	GOOS=linux GOARCH=arm go build -o $(BIN_NAME)
build: clean
	go build -o $(BIN_NAME)

format:
	gofmt -w $(ALL_GO_FILES)

test:
	@echo "Running test"
	go test -v ./...

setup:
	@echo "Starting lignum in docker"
	docker-compose up -d

run:
	@go run main.go -config $(CONFIG_FILE)
