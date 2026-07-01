//go:build mage

package main

import (
	"bufio"
	"os"
	"runtime"
	"strings"

	"github.com/magefile/mage/sh"
)

const localEnvFile = ".env.local"

func Deps() error {
	if err := sh.RunV("go", "mod", "download"); err != nil {
		return err
	}
	return sh.RunV("go", "mod", "tidy")
}

func Generate() error {
	return sh.RunV("go", "generate", "./...")
}

func Build() error {
	return sh.RunV("go", "build", "-o", binaryName(), "./cmd/scheduler")
}

func Run() error {
	if err := loadLocalEnv(); err != nil {
		return err
	}
	return sh.RunV("go", "run", "./cmd/scheduler", "app")
}

func Test() error {
	return sh.RunV("go", "test", "-v", "./...")
}

func TestCoverage() error {
	if err := sh.RunV("go", "test", "-v", "-coverprofile=coverage.out", "./..."); err != nil {
		return err
	}
	return sh.RunV("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html")
}

func Lint() error {
	return sh.RunV("golangci-lint", "run")
}

func DockerBuild() error {
	return sh.RunV("docker", "build", "-t", "scheduler:latest", ".")
}

func Migrate() error {
	if err := loadLocalEnv(); err != nil {
		return err
	}
	return sh.RunV("go", "run", "./cmd/scheduler", "migrate")
}

func Clean() error {
	for _, path := range []string{"scheduler", "scheduler.exe", "coverage.out", "coverage.html"} {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return sh.RunV("go", "clean")
}

func loadLocalEnv() error {
	file, err := os.Open(localEnvFile)
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
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
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

func binaryName() string {
	if runtime.GOOS == "windows" {
		return "scheduler.exe"
	}
	return "scheduler"
}
