package main

import (
	"fmt"
	"io"
	"os"
)

type cd struct{}

func (c cd) Run(stdout, stderr io.Writer, args ...string) error {
	if len(args) > 1 {
		return fmt.Errorf("too many arguments")
	}

	if len(args) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		return os.Chdir(home)
	}

	path := args[0]
	return os.Chdir(path)
}

type exit struct{}

func (e exit) Run(stdout, stderr io.Writer, args ...string) error {
	os.Exit(0)
	return nil
}
