BZL_PB_DIR = bzl/bzlpb
# XXX(kormat): this doesn't need to precisely match the version of bazel you're
# using. The protobuf doesn't change that often/much, and breaking changes are
# likely to be rare.
BZL_VERSION = 0.26.1
BZL_BUILD_PROTO_URL = https://github.com/bazelbuild/bazel/raw/$(BZL_VERSION)/src/main/protobuf/build.proto

GIT_VERSION = $(shell git describe --tag --dirty)
OS = $(shell uname -s | tr '[:upper:]' '[:lower:]')
MACH = $(shell uname -m)
LDFLAGS = -ldflags "-X main.version=${GIT_VERSION}"

all: binary

.PHONY: binary
binary:
	go build ${LDFLAGS} -o bzlcompat

.PHONY: release
release:
	fn="bzlcompat-$(GIT_VERSION)-$(OS)-$(MACH)"; \
		go build ${LDFLAGS} -o "$$fn"; \
		sha256sum "$$fn" | tee "$$fn.sha256"

.PHONY: proto
proto:
	if [ ! -e $(BZL_PB_DIR)/build.proto ]; then mkdir -p $(BZL_PB_DIR) \
		&& cd $(BZL_PB_DIR) && curl -sSLO $(BZL_BUILD_PROTO_URL); fi
	protoc --go_out=. $(BZL_PB_DIR)/build.proto
