package bzl

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/golang/protobuf/proto"

	bzlpb "github.com/kormat/bzlcompat/bzl/bzlpb"
)

var ruleRx = regexp.MustCompile("^@(?P<repository>[^/]+)//(?P<subdir>[^:]*):.*$")

// ExtGoLib represents an external go library.
type ExtGoLib struct {
	ImportPath string
}

// LoadQuery parses the output of `bazel query` and returns a map of go external
// repos to their import paths.
func LoadQuery(b []byte) (map[string]ExtGoLib, error) {
	exts := make(map[string]ExtGoLib)
	qr := &bzlpb.QueryResult{}
	if err := proto.Unmarshal(b, qr); err != nil {
		return nil, err
	}
	for _, t := range qr.Target {
		if *t.Type != bzlpb.Target_RULE {
			continue
		}
		if (*t.Rule.Name)[0] != '@' {
			continue
		}
		if *t.Rule.RuleClass != "go_library" {
			continue
		}
		matches := ruleRx.FindStringSubmatch(*t.Rule.Name)
		name := matches[1]
		subdir := matches[2]
		if strings.HasPrefix(subdir, "vendor/") {
			// Skip vendored code in external repos.
			continue
		}
		if _, ok := exts[name]; ok {
			// External repo is already in map.
			continue
		}
		importPath := ""
		for _, attr := range t.Rule.Attribute {
			if *attr.Name != "importpath" {
				continue
			}
			if *attr.Type != bzlpb.Attribute_STRING {
				return nil, fmt.Errorf("Rule attribute 'importpath' is not of type STRING")
			}
			importPath = *attr.StringValue
		}
		if importPath == "" {
			return nil, fmt.Errorf("Unable to find importpath for %s", *t.Rule.Name)
		}
		if len(subdir) > 0 {
			importPath = strings.TrimSuffix(importPath, "/"+subdir)
		}
		exts[name] = ExtGoLib{ImportPath: importPath}
	}
	return exts, nil
}
