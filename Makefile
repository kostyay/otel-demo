
.PHONY: setup
setup:
	@echo "Setting up environment..."
	@go install github.com/bufbuild/buf/cmd/buf@latest
	@go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install github.com/bufbuild/connect-go/cmd/protoc-gen-connect-go@latest
	@echo "Done."

.PHONY: controller
controller:
	@echo "Generating controller..."
	@buf generate
	@echo "Done."