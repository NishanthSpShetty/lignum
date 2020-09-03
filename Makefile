BIN_NAME=./bin/lignum
ALL_GO_FILES=$(shell find . -type f  -name '*.go')
CONFIG_FILE="config.yml"

clean:
	@echo $(ALL_GO_FILES)
	@test ! -e bin || rm -r bin

build: clean
	go build -o $(BIN_NAME)

format:
	gofmt -w $(ALL_GO_FILES)
run1:
	@echo "Running single instance on the host..."
	go run main.go -config $(CONFIG_FILE)
