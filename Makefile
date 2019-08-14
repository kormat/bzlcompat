BZL_PB_DIR = bzl/bzlpb
# XXX(kormat): this doesn't need to precisely match the version of bazel you're
# using. The protobuf doesn't change that often/much, and breaking changes are
# likely to be rare.
BZL_VERSION = 0.26.1
BZL_BUILD_PROTO_URL = https://github.com/bazelbuild/bazel/raw/$(BZL_VERSION)/src/main/protobuf/build.proto

.PHONY: proto
proto:
	if [ ! -e $(BZL_PB_DIR)/build.proto ]; then mkdir -p $(BZL_PB_DIR) \
		&& cd $(BZL_PB_DIR) && curl -sSLO $(BZL_PB_DIR); fi
	protoc --go_out=. $(BZL_PB_DIR)/build.proto

.PHONY: release
release:
	@ver=$$(git describe --tag)$$(git diff HEAD --quiet || echo '-dirty'); \
		os=$$(uname -s | tr '[:upper:]' '[:lower:]'); \
		mach=$$(uname -m); \
		fn="bzlcompat-$$ver-$$os-$$mach"; \
		go build -o "$$fn"; sha256sum "$$fn" | tee "$$fn.sha256"
