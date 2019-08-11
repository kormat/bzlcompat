package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"github.com/kormat/bzlcompat/bzl"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	info, err := getBzlInfo()
	if err != nil {
		log.Fatalf("FATAL: %s", err)
		os.Exit(1)
	}
	log.Printf("Bazel version: %s", info.Version)
	exts, err := getExtGoDeps()
	if err != nil {
		log.Fatalf("FATAL: %s", err)
		os.Exit(1)
	}
	log.Printf("Found %d external dependencies", len(exts))
	count, err := makeLinks(info, exts)
	if err != nil {
		log.Fatalf("FATAL: %s", err)
		os.Exit(1)
	}
	log.Printf("Created %d symlinks in vendor/", count)
}

func getBzlInfo() (*bzl.Info, error) {
	cmd := exec.Command("bazel", "info")
	b, err := runCmd(cmd)
	if err != nil {
		return nil, err
	}
	return bzl.InfoFromString(string(b))
}

func getExtGoDeps() (map[string]bzl.ExtGoLib, error) {
	cmd := exec.Command("bazel", "query", "kind('go_repository rule', //external:*)", "--output=proto")
	b, err := runCmd(cmd)
	if err != nil {
		return nil, err
	}
	return bzl.LoadGoQuery(b)
}

func runCmd(cmd *exec.Cmd) ([]byte, error) {
	b, err := cmd.Output()
	if err != nil {
		switch err := err.(type) {
		case *exec.ExitError:
			status := err.ProcessState.Sys().(syscall.WaitStatus)
			return b, fmt.Errorf("'%s' exited with %d:\n\n%s",
				strings.Join(cmd.Args, " "), status.ExitStatus(), string(err.Stderr))
		}
	}
	return b, nil
}

func makeLinks(info *bzl.Info, exts map[string]bzl.ExtGoLib) (int, error) {
	count := 0
	for k, v := range exts {
		src := path.Join("vendor", v.ImportPath)
		if err := os.MkdirAll(path.Dir(src), os.ModePerm); err != nil {
			return 0, fmt.Errorf("unable to create dir: %v", err)
		}
		dest := path.Join(info.OutputBase, "external", k)
		if created, err := makeLink(src, dest); err != nil {
			return 0, err
		} else if created {
			count++
		}
	}
	return count, nil
}

func makeLink(src, dest string) (bool, error) {
	fi, err := os.Lstat(src)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	if err == nil {
		// File exists, ensure it's a symlink
		if fi.Mode()&os.ModeSymlink == 0 {
			return false, fmt.Errorf("non-symlink in the way: %s", src)
		}
		origDest, err := os.Readlink(src)
		if err != nil {
			return false, fmt.Errorf("unable to read existing symlink: %v", err)
		}
		// Symlink already points to the correct place, no need to touch it.
		if origDest == dest {
			return false, nil
		}
		if err := os.Remove(src); err != nil {
			return false, err
		}
	}
	if err = os.Symlink(dest, src); err != nil {
		return false, fmt.Errorf("unable to create symlink: %v", err)
	}
	return true, nil
}
