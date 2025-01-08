package main

import (
	"errors"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/manifoldco/promptui"
)

const (
	version = "v1.0.0"
)

type Release mg.Namespace

func (Release) Publish() error {
	err := createTag(version)
	if err != nil {
		return err
	}

	token, err := resolveToken()
	if err != nil {
		return err
	}

	env := map[string]string{
		"GITHUB_TOKEN": token,
	}

	publish, err := confirm("Publish")
	if err != nil {
		return err
	}

	opts := []string{"release"}
	if !publish {
		opts = append(
			opts,
			"--skip=publish",
			"--auto-snapshot",
		)
	}

	return sh.RunWithV(
		env,
		"goreleaser",
		opts...,
	)
}

func (Release) Clean() error {
	return sh.Rm("dist")
}

func createTag(v string) error {
	err := sh.RunV("git", "tag", v, "-f")
	if err != nil {
		return err
	}
	return sh.RunV("git", "push", "origin", v)
}

func confirm(label string) (bool, error) {
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
	}
	_, err := prompt.Run()
	if err != nil {
		if errors.Is(err, promptui.ErrAbort) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func resolveToken() (string, error) {
	return sh.Output("gostore", "cat", "github.com/release-token")
}
