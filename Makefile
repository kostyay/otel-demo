
.PHONY: setup
setup:
	@echo "Setting up environment..."
	@go install github.com/bufbuild/buf/cmd/buf@latest
	@go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install github.com/bufbuild/connect-go/cmd/protoc-gen-connect-go@latest
	@echo "Done."

.PHONY: api
api:
	@echo "Generating controller..."
	@buf generate
	@echo "Done."


.PHONY: fmt
fmt:
	goimports -w $$(find . -name "*.go" | grep -v -E '(.git|/vendor/|pb.go|protodep|grpcmock_cmds)')
