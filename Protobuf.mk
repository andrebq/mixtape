PROTO_DIR := api
PROTO_FILES := $(wildcard $(PROTO_DIR)/*.proto)

# eventually, replace this with the new go tools module
install_tooling:
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest


generate-proto: install_protoc install_plugins
	@cd ./proto && buf lint
	@cd ./proto && buf generate

.PHONY: install_protoc install_plugins generate
