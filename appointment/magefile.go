//go:build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

const openAPIGenerator = "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.7.1"

func Generate() error {
	if err := run("go", "run", "github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0", "generate"); err != nil {
		return err
	}
	if err := run(
		"go", "run", openAPIGenerator,
		"--config", "internal/port/http/client/customer/customer-client.oapi-codegen.yaml",
		"-o", "internal/port/http/client/customer/customer.gen.go",
		"api-spec/customer-api.yaml",
	); err != nil {
		return err
	}
	return nil
}
func Build() error {
	if err := Generate(); err != nil {
		return err
	}
	if err := os.MkdirAll("build", 0755); err != nil {
		return err
	}
	return run("go", "build", "-o", binaryPath(), "./cmd/appointment")
}

func Test() error {
	return run("go", "test", "./...")
}

func Lint() error {
	if err := run("go", "fmt", "./..."); err != nil {
		return err
	}
	return run("go", "vet", "./...")
}

func Check() error {
	if err := Generate(); err != nil {
		return err
	}
	if err := Lint(); err != nil {
		return err
	}
	return Test()
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %v: %w", name, args, err)
	}
	return nil
}

func binaryPath() string {
	if runtime.GOOS == "windows" {
		return "build/appointment.exe"
	}
	return "build/appointment"
}
