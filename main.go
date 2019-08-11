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
	count := 0
	for k, v := range exts {
		fullPath := path.Join("vendor", v.ImportPath)
		dir := path.Dir(fullPath)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			log.Fatalf("FATAL: unable to create dir: %v", err)
		}
		if err := os.Symlink(path.Join(info.OutputBase, "external", k), fullPath); err != nil {
			if !os.IsExist(err) {
				log.Fatalf("FATAL: unable to create symlink: %v", err)
			}
		} else {
			count++
		}
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
