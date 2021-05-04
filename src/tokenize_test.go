package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type spyCommand struct {
	executed  bool
	arguments []string

	stdout string
	stderr string

	err error
}

func (sc *spyCommand) Run(stdout, stderr io.Writer, args ...string) error {
	sc.executed = true
	sc.arguments = args

	if sc.stdout != "" {
		if _, err := stdout.Write([]byte(sc.stdout)); err != nil {
			return fmt.Errorf("when writing stdout: %w", err)
		}

	}

	if sc.stderr != "" {
		if _, err := stderr.Write([]byte(sc.stderr)); err != nil {
			return fmt.Errorf("when writing stderr: %w", err)
		}

	}

	return sc.err
}

func TestCommandRan(t *testing.T) {
	sc := &spyCommand{
		stdout: "spy ran",
		stderr: "but there was also a warning",
	}

	var cf CommandFinder = &MockCommandFinder{
		commands: map[string]Command{
			"spy": sc,
		},
	}

	var actualStdout bytes.Buffer
	var actualStderr bytes.Buffer

	input := "spy hello"

	s := NewShell(
		WithCommandFinder(cf),
		NonInteractive(input),
		WithStderr(&actualStderr),
		WithStdout(&actualStdout),
	)

	s.Run()

	if !sc.executed {
		t.Errorf("spy command never ran")
	}

	expectedArgs := []string{"hello"}
	if diff := cmp.Diff(expectedArgs, sc.arguments); diff != "" {
		t.Errorf("unexpected spy arguments (-got +want)\n%s", diff)
	}

	assertOutput(t, "stdout", "spy ran", actualStdout)
	assertOutput(t, "stderr", "but there was also a warning", actualStderr)
}

func assertOutput(t *testing.T, name, expected string, actual bytes.Buffer) {
	t.Helper()

	if diff := cmp.Diff(expected, actual.String()); diff != "" {
		t.Errorf("unexpected %s output (-want +got)\n%s", name, diff)
	}
}

type MockCommandFinder struct {
	commands map[string]Command
}

func (cf *MockCommandFinder) Lookup(name string) (Command, error) {
	c, present := cf.commands[name]
	if !present {
		return nil, &CommandNotFoundError{name: name}
	}

	return c, nil
}
