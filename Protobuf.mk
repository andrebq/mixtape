PROTO_DIR := api
PROTO_FILES := $(wildcard $(PROTO_DIR)/*.proto)
PROTOC_VERSION := 24.4
PROTOC_BIN := /usr/local/bin/protoc

# Determine system architecture
ARCH := $(shell uname -m)
ifeq ($(ARCH),arm64)
	PROTOC_ARCH := osx-aarch_64
else
	PROTOC_ARCH := osx-x86_64
endif

PROTOC_URL := https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-$(PROTOC_ARCH).zip
PROTOC_ZIP := /tmp/protoc-$(PROTOC_VERSION).zip

# Install protoc if not present
install_protoc:
	@if ! command -v protoc >/dev/null 2>&1; then \
		echo "protoc not found. Downloading..."; \
		curl -L $(PROTOC_URL) -o $(PROTOC_ZIP); \
		unzip -o $(PROTOC_ZIP) -d /tmp/protoc; \
		mkdir -p /usr/local/include/google/protobuf; \
		cp -r /tmp/protoc/include/* /usr/local/include/; \
		cp /tmp/protoc/bin/protoc $(PROTOC_BIN); \
		chmod +x $(PROTOC_BIN); \
		rm -rf /tmp/protoc $(PROTOC_ZIP); \
		echo "protoc installed successfully."; \
	else \
		echo "protoc is already installed."; \
	fi

# Ensure required Go plugins are installed
install_plugins:
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

generate-proto: install_protoc install_plugins
	@protoc --proto_path=$(PROTO_DIR) \
		--go_out=$(PROTO_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)

.PHONY: install_protoc install_plugins generate
