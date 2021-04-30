package main

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEchoInteractive(t *testing.T) {
	t.Skip("this is disabled - only useful as a motivating example")
	inputs := []string{"hello", "my", "name", "is", "geoffrey"}

	expected := []string{PROMPT}
	for _, input := range inputs {
		expected = append(expected, []string{input, "\n", PROMPT}...)
	}

	var actual bytes.Buffer

	s := Shell{
		interactive: true,
	}

	s.Run(strings.NewReader(strings.Join(inputs, "\n")), &actual)

	if diff := cmp.Diff(strings.Join(expected, ""), actual.String()); diff != "" {
		t.Errorf("non-zero diff (-want +got):\n%s", diff)
	}

}

func TestEchoNonInteractive(t *testing.T) {
	t.Skip("this is disabled - only useful as a motivating example")
	input := "hello"
	expected := "hello\n"

	var actual bytes.Buffer

	s := Shell{}

	s.Run(strings.NewReader(input), &actual)

	if diff := cmp.Diff(expected, actual.String()); diff != "" {
		t.Errorf("non-zero diff (-want +got):\n%s", diff)
	}

}

type spyCommand struct {
	executed  bool
	arguments []string

	output string
	err    error
}

func (sc *spyCommand) Run(out io.Writer, args ...string) error {
	sc.executed = true
	sc.arguments = args

	w := bufio.NewWriter(out)
	w.WriteString(sc.output)
	w.Flush()

	return sc.err
}

// This really is a builtin....
// Not sure if it's worth unit testing the fork + exec process explicitly - I could mock a strategy for finding non-builtins though...
func TestCommandRan(t *testing.T) {
	sc := &spyCommand{
		output: "spy ran",
	}

	var cf CommandFinder = &MockCommandFinder{
		commands: map[string]Command{
			"spy": sc,
		},
	}

	s := Shell{
		cf: cf,
	}

	var output bytes.Buffer

	s.Run(strings.NewReader("spy hello"), &output)

	if !sc.executed {
		t.Errorf("spy command never ran")
	}

	if diff := cmp.Diff([]string{"spy", "hello"}, sc.arguments); diff != "" {
		t.Errorf("unexpected spy arguments (-got +want)\n%s", diff)
	}

	if diff := cmp.Diff("spy ran\n", output.String()); diff != "" {
		t.Errorf("unexpected output (-want +got)\n%s", diff)
	}

}

type MockCommandFinder struct {
	commands map[string]Command
}

func (cf *MockCommandFinder) Lookup(name string) (Command, error) {
	c, present := cf.commands[name]
	if !present {
		return nil, CommandMissingError{name}
	}

	return c, nil
}
