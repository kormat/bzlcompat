package bzl

import (
	"fmt"
	"strings"
)

// Info contains the relevant parts of the output of 'bazel info':
//   release: release 0.26.1
//   output_base: /home/user/.cache/bazel/_bazel_user/60ed071115454b1cf2fea18770779bb0
//   workspace: /home/user/go/src/github.com/user/repo
type Info struct {
	Version    string
	OutputBase string
	Workspace  string
}

// InfoFromString parses the output of `bazel info`, and returns an *Info.
func InfoFromString(s string) (*Info, error) {
	var info Info
	for _, line := range strings.Split(s, "\n") {
		words := strings.Split(line, " ")
		switch words[0] {
		case "release:":
			// Be a little paranoid here, just in case the format changes
			if len(words) != 3 || words[1] != "release" {
				return nil, fmt.Errorf(
					"Unable to parse version from 'bazel info' output:\n    %v", line)
			}
			info.Version = words[2]
		case "output_base:":
			info.OutputBase = words[1]
		case "workspace:":
			info.Workspace = words[1]
		}
	}
	if info.Version == "" {
		return nil, fmt.Errorf("Unable to find 'release:' line in 'bazel info' output")
	} else if info.OutputBase == "" {
		return nil, fmt.Errorf("Unable to find 'output_base:' line in 'bazel info' output")
	} else if info.Workspace == "" {
		return nil, fmt.Errorf("Unable to find 'workspace:' line in 'bazel info' output")
	}
	return &info, nil
}

func (info *Info) String() string {
	return fmt.Sprintf("Version: %s\nOutputBase: %s\nWorkspace: %s\n",
		info.Version, info.OutputBase, info.Workspace)
}
