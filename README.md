# bzlcompat
Because using bazel doesn't have to break your normal tools. `bzlcompat` queries
bazel for all external dependencies of your codebase, and creates a symlink farm
so your normal tools can find/use them.

(**N.B.**: currently `bzlcompat` only supports Go, but supporting python should be
easy, for example.)

## Installation
Download binaries from https://github.com/kormat/bzlcompat/releases, or built it
yourself following the instructions in the #development section below.

## Usage
Add `vendor/` to `.bazelignore` in the top-level of your workspace.

Run `bzlcompat` inside a bazel workspace. It will query bazel, and
create a symlink farm under `vendor/` for all external go dependencies.


## Development:
To build `bzlcompat` yourself, install these dependencies:
```
sudo apt-get install curl protobuf-compiler
go get -u github.com/golang/protobuf/protoc-gen-go
```
and then run `make proto`. This will download `build.proto` from the bazel
project, and generate `build.pb.go` in `bzl/bzlpb`. This is used for
`bzlcompat` to understand the output of `bazel query`.

After that, `go build` will give you a `bzlcompat` binary, which you can copy to
wherever you need it to go.
