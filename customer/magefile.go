//go:build mage

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func Generate() error {
	if err := run("go", "run", "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.7.1", "--config", "internal/port/http/server/customer/oapi-codegen.yaml", "-o", "internal/port/http/server/customer/openapi.gen.go", "api-spec/customer-api.yaml"); err != nil {
		return err
	}
	if err := run("go", "run", "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.7.1", "--config", "internal/port/http/server/fidelity/oapi-codegen.yaml", "-o", "internal/port/http/server/fidelity/openapi.gen.go", "api-spec/fidelity-card-api.yaml"); err != nil {
		return err
	}
	return run("go", "run", "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.7.1", "--config", "internal/port/http/server/wallet/oapi-codegen.yaml", "-o", "internal/port/http/server/wallet/openapi.gen.go", "api-spec/wallet-api.yaml")
}

func Build() error {
	if err := os.MkdirAll("build", 0755); err != nil {
		return err
	}
	return run("go", "build", "-o", binaryPath(), "./cmd/customer")
}

func Run() error {
	if err := loadLocalEnv(); err != nil {
		return err
	}
	return run("go", "run", "./cmd/customer", "app")
}

func Migrate() error {
	if err := loadLocalEnv(); err != nil {
		return err
	}
	return run("go", "run", "./cmd/customer", "migrate", "up")
}

func Test() error { return run("go", "test", "./...") }

func Check() error {
	if err := Generate(); err != nil {
		return err
	}
	if err := run("go", "fmt", "./..."); err != nil {
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

func loadLocalEnv() error {
	file, err := os.Open(".env.local")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func binaryPath() string {
	if runtime.GOOS == "windows" {
		return "build/customer.exe"
	}
	return "build/customer"
}
