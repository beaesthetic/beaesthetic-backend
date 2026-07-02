//go:build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func Generate() error {
	if err := run("go", "run", "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest", "-generate", "types,gin-server", "-package", "api", "-o", "internal/api/notification.gen.go", "api-spec/notification-api.yaml"); err != nil {
		return err
	}
	return run("go", "run", "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest", "-generate", "types,gin-server", "-package", "smswebhook", "-o", "internal/api/smswebhook/sms_webhook.gen.go", "api-spec/sms-gateway-webhook.yaml")
}

func Build() error {
	if err := os.MkdirAll("build", 0755); err != nil {
		return err
	}
	return run("go", "build", "-o", binaryPath(), "./cmd/notification")
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
		return "build/notification.exe"
	}
	return "build/notification"
}

