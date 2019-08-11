package bzl

import (
	"fmt"
	"log"

	"github.com/golang/protobuf/proto"

	bzlpb "github.com/kormat/bzlcompat/bzl/bzlpb"
)

// ExtGoLib represents an external go library.
type ExtGoLib struct {
	ImportPath string
}

// LoadGoQuery parses the output of `bazel query` and returns a map of go external
// repos to their import paths. Example (in XML format):
//
// <rule class="go_repository" location="/home/user/go/src/github.com/user/repo/WORKSPACE:146:1" name="//external:com_github_axw_gocov">
//   <string name="name" value="com_github_axw_gocov"/>
//   <string name="importpath" value="github.com/axw/gocov"/>
//   <string name="commit" value="54b98cfcac0c63fb3f9bd8e7ad241b724d4e985b"/>
// </rule>
//
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
		if *t.Rule.RuleClass != "go_repository" {
			log.Printf("WARN: expected rule of class go_repository, instead got %s", *t.Rule.RuleClass)
			continue
		}
		var name, importPath string
		for _, attr := range t.Rule.Attribute {
			switch *attr.Name {
			case "name":
				name, err = readStringAttr(attr)
				if err != nil {
					return nil, err
				}
			case "importpath":
				importPath, err = readStringAttr(attr)
				if err != nil {
					return nil, err
				}
			}
		}
		if name == "" {
			return nil, fmt.Errorf("Unable to find name for %s", *t.Rule.Name)
		}
		if importPath == "" {
			return nil, fmt.Errorf("Unable to find importpath for %s", *t.Rule.Name)
		}
		exts[name] = ExtGoLib{ImportPath: importPath}
	}
	return exts, nil
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
