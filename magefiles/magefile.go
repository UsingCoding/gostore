package main

import (
	"context"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

//nolint:gochecknoinits
func init() {
	// enable verbose explicitly
	_ = os.Setenv(mg.VerboseEnv, "1")
}

const (
	appID = "gostore"
)

var Default = All

func All(ctx context.Context) {
	mg.Verbose()
	mg.SerialCtxDeps(ctx, Modules, Build)
	mg.CtxDeps(ctx, Lint, Test)
}

func Build(_ context.Context) error {
	return sh.RunV(
		"go",
		"build",
		"-v",
		"-o",
		"./bin/"+appID,
		"./cmd/"+appID,
	)
}

func Modules() error {
	return sh.RunV("go", "mod", "tidy")
}

func Test() error {
	return sh.RunV("go", "test", "./...")
}

func Lint() error {
	return sh.RunV("golangci-lint", "run", "--timeout", "5m")
}
