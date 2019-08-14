package bzl

import (
	"fmt"
	"log"
	"strings"

	"github.com/golang/protobuf/proto"

	bzlpb "github.com/kormat/bzlcompat/bzl/bzlpb"
)

// ExtGoLib represents an external go library.
type ExtGoLib struct {
	ImportPath string
}

// LoadGoQuery parses the output of `bazel query` and returns a map of go external
// repos to their import paths.
func LoadGoQuery(b []byte) (map[string]ExtGoLib, error) {
	exts := make(map[string]ExtGoLib)
	qr := &bzlpb.QueryResult{}
	if err := proto.Unmarshal(b, qr); err != nil {
		return nil, err
	}
	for _, t := range qr.Target {
		var err error
		if *t.Type != bzlpb.Target_RULE {
			log.Printf("WARN: expected target of type RULE, instead got %s", *t.Type)
			continue
		}
		var name, importPath string
		switch *t.Rule.RuleClass {
		case "go_repository":
			if name, importPath, err = parseGoRepository(t.Rule); err != nil {
				return nil, err
			}
		case "git_repository":
			if name, importPath, err = parseGitRepository(t.Rule); err != nil {
				return nil, err
			}
		default:
			log.Printf("WARN: expected rule of class [go_repository, git_repository], "+
				"instead got %s", *t.Rule.RuleClass)
			continue
		}
		switch name {
		case "":
			return nil, fmt.Errorf("Unable to find name for %s", *t.Rule.Name)
		case "org_golang_google_grpc":
			// XXX(kormat): workaround for mismatch between url and import path:
			//   https://github.com/bazelbuild/rules_go/blob/0.18.7/go/private/repositories.bzl#L164
			importPath = "google.golang.org/grpc"
		case "org_golang_google_genproto":
			// XXX(kormat): workaround for mismatch between url and import path:
			//   https://github.com/bazelbuild/rules_go/blob/0.18.7/go/private/repositories.bzl#L178
			importPath = "google.golang.org/genproto"
		}
		if importPath == "" {
			return nil, fmt.Errorf("Unable to find importpath for %s", *t.Rule.Name)
		}
		exts[name] = ExtGoLib{ImportPath: importPath}
	}
	return exts, nil
}

// <rule class="go_repository" location="/home/user/go/src/github.com/user/repo/WORKSPACE:146:1" name="//external:com_github_axw_gocov">
//   <string name="name" value="com_github_axw_gocov"/>
//   <string name="importpath" value="github.com/axw/gocov"/>
//   <string name="commit" value="54b98cfcac0c63fb3f9bd8e7ad241b724d4e985b"/>
// </rule>
func parseGoRepository(r *bzlpb.Rule) (string, string, error) {
	var name, importPath string
	var err error
	for _, attr := range r.Attribute {
		switch *attr.Name {
		case "name":
			name, err = readStringAttr(attr)
			if err != nil {
				return "", "", err
			}
		case "importpath":
			importPath, err = readStringAttr(attr)
			if err != nil {
				return "", "", err
			}
		}
	}
	return name, importPath, nil
}

// As there's no (easy?) way to determine if a git_repository rule contains go source,
// treat them as a go_repository just in case.
//
// <?xml version="1.1" encoding="UTF-8" standalone="no"?>
// <query version="2">
//     <rule class="git_repository" location="/home/user/.cache/bazel/_bazel_user/5fc5ba52f0a32618e694c19b268efbf4/external/io_bazel_rules_go/go/private/repositories.bzl:237:9" name="//external:com_github_golang_protobuf">
//         <string name="name" value="com_github_golang_protobuf"/>
//         <string name="remote" value="https://github.com/golang/protobuf"/>
//         <string name="commit" value="c823c79ea1570fb5ff454033735a8e68575d1d0f"/>
//         <string name="shallow_since" value="1549405252 -0800"/>
//         <list name="patches">
//             <label value="@io_bazel_rules_go//third_party:com_github_golang_protobuf-gazelle.patch"/>
//             <label value="@io_bazel_rules_go//third_party:com_github_golang_protobuf-extras.patch"/>
//         </list>
//         <list name="patch_args">
//             <string value="-p1"/>
//         </list>
//         <rule-input name="@io_bazel_rules_go//third_party:com_github_golang_protobuf-extras.patch"/>
//         <rule-input name="@io_bazel_rules_go//third_party:com_github_golang_protobuf-gazelle.patch"/>
//     </rule>
// </query>
func parseGitRepository(r *bzlpb.Rule) (string, string, error) {
	var name, remote string
	var err error
	for _, attr := range r.Attribute {
		switch *attr.Name {
		case "name":
			name, err = readStringAttr(attr)
			if err != nil {
				return "", "", err
			}
		case "remote":
			remote, err = readStringAttr(attr)
			if err != nil {
				return "", "", err
			}
		}
	}
	if strings.Contains(remote, "://") {
		remote = strings.Split(remote, "://")[1]
	}
	return name, remote, nil
}

func readStringAttr(a *bzlpb.Attribute) (string, error) {
	if err := checkAttrType(a, bzlpb.Attribute_STRING); err != nil {
		return "", err
	}
	return *a.StringValue, nil
}

func checkAttrType(a *bzlpb.Attribute, _type bzlpb.Attribute_Discriminator) error {
	if *a.Type != _type {
		return fmt.Errorf("Rule attribute %s is not of type %s, instead got %s",
			*a.Name, _type, *a.Type)
	}
	return nil
}
