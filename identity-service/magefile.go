//go:build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func Build() error {
	if err := os.MkdirAll("build", 0o755); err != nil {
		return err
	}
	return run("go", "build", "-o", binaryPath(), "./cmd/identity")
}

func Test() error { return run("go", "test", "./...") }

func Lint() error {
	if err := run("go", "fmt", "./..."); err != nil {
		return err
	}
	return run("go", "vet", "./...")
}

func Check() error {
	if err := Lint(); err != nil {
		return err
	}
	return Test()
}

func run(name string, args ...string) error {
	command := exec.Command(name, args...)
	command.Stdout, command.Stderr, command.Env = os.Stdout, os.Stderr, os.Environ()
	if err := command.Run(); err != nil {
		return fmt.Errorf("%s %v: %w", name, args, err)
	}
	return nil
}

func binaryPath() string {
	if runtime.GOOS == "windows" {
		return "build/identity.exe"
	}
	return "build/identity"
}
