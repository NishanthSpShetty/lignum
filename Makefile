BIN_NAME=./bin/lignum
ALL_GO_FILES=$(shell find . -type f  -name '*.go')
CONFIG_FILE="config.yml"
LCONFIG_FILE="configleader.yml"
export ENV=development

.PHONY: test proto

clean:
	@test ! -e bin || rm -r bin

build-linux:clean
	GOOS=linux GOARCH=arm go build -o $(BIN_NAME)
build: clean
	go build -o $(BIN_NAME)

fmt:
	gofumpt -w $(ALL_GO_FILES)

test:
	@echo "Running test"
	go test -v ./...

setup:
	@echo "Starting lignum in docker"
	docker-compose up -d

runl:
	@go run main.go -config $(LCONFIG_FILE)

run:
	@go run main.go -config $(CONFIG_FILE)
lint:
	golint ./...

proto:
	@echo "generating grpc code"
	@protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	proto/lignum.proto
