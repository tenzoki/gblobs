BIN=cmd/gblobs/gblobs

.PHONY: all build test demo clean help

all: build test

build:
	@go build -o $(BIN) ./cmd/gblobs
	@echo "Build complete: CLI binary available at ./cmd/gblobs/gblobs"

test:
	@go test ./test/...

demo: build
	@echo "\n=== Running Go API demos ==="
	go run demo/simple_module.go
	go run demo/encrypted_demo.go
	go run demo/dedup_demo.go
	@echo "\n=== Running CLI demo ==="
	cd demo && bash cli_walkthrough.sh
	@echo "\n=== Running Search demo ==="
	cd demo && bash search_demo.sh

clean:
	rm -rf cmd/gblobs/gblobs demo/tmp_* demo/tmp_cli_gblobs_store demo/tmp_gblobs_encstore demo/tmp_gblobs_dedup demo/demo_file.txt tmp_* demo_file.txt demo/ml_doc.txt demo/meeting.txt demo/recipe.txt demo/shopping.txt

help:
	@echo "Targets:"
	@echo "  build      Build the CLI binary ($(BIN))"
	@echo "  test       Run all tests (go test ./...)"
	@echo "  demo       Run Go, CLI, and search demos in ./demo/"
	@echo "  clean      Remove build/demo artifacts"
	@echo "  install    Run go mod tidy (dependency tidy up)"
	@echo "  help       Show this help text"
.PHONY: all build test demo clean help install

install:
	@go mod tidy
