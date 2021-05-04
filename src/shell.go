package main

import (
	"io"
	"os"
	"strings"
)

const PROMPT = "(g-shell)> "

type shell struct {
	interactive bool

	cf CommandFinder

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func NewShell(opts ...ShellOption) *shell {

	s := &shell{
		cf: &CommandStore{
			builtins: map[string]Command{
				"cd":   cd{},
				"exit": exit{},
			},
		},
		interactive: true,

		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *shell) PrintPrompt() {
	if s.interactive {
		s.Stdout.Write([]byte(PROMPT))
	}
}

type ShellOption func(*shell)

func NonInteractive(command string) ShellOption {
	return func(s *shell) {
		s.interactive = false
		s.Stdin = strings.NewReader(command)
	}
}

func WithCommandFinder(cf CommandFinder) ShellOption {
	return func(s *shell) {
		s.cf = cf
	}
}

func WithStdout(w io.Writer) ShellOption {
	return func(s *shell) {
		s.Stdout = w
	}
}

func WithStderr(w io.Writer) ShellOption {
	return func(s *shell) {
		s.Stderr = w
	}
}
