package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

const PROMPT = "(g-shell)> "

var command string

type Shell struct {
	interactive bool

	cf CommandFinder
}

func main() {
	flag.StringVar(&command, "c", "", command)
	flag.Parse()

	s := Shell{
		interactive: true,
		cf:          &CommandStore{},
	}

	var from io.Reader = os.Stdin
	var to io.Writer = os.Stdout

	if isFlagPassed("c") {
		s.interactive = false
		from = strings.NewReader(command)
	}

	s.Run(from, to)
}

func (s *Shell) Run(r io.Reader, w io.Writer) {
	var wg sync.WaitGroup

	inputs := make(chan string)
	outputs := make(chan string)

	wg.Add(1)
	go func() {
		defer wg.Done()

		s.runCommand(inputs, outputs)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		s.output(w, outputs)
	}()

	s.tokenize(r, inputs)
	wg.Wait()
}

func (s *Shell) output(to io.Writer, outputs chan string) {
	w := bufio.NewWriter(to)

	if s.interactive {
		fmt.Fprint(w, PROMPT)
		w.Flush()
	}

	for out := range outputs {
		fmt.Fprintln(w, out)
		w.Flush()

		if s.interactive {
			fmt.Fprint(w, PROMPT)
			w.Flush()
		}
	}
}

func (s *Shell) runCommand(inputs, outputs chan string) {
	for input := range inputs {
		fields := strings.Fields(input)

		if len(fields) == 0 {
			outputs <- ""
			continue
		}

		cmd := fields[0]
		rest := fields[1:]

		c, err := s.cf.Lookup(cmd)
		if err != nil {
			outputs <- err.Error()
			continue
		}

		var w bytes.Buffer

		args := append([]string{cmd}, rest...)

		err = c.Run(&w, args...)
		if err != nil {
			outputs <- err.Error()
			continue
		}

		outputs <- w.String()
	}

	close(outputs)
}

func (s *Shell) tokenize(r io.Reader, to chan string) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		input := scanner.Text()
		to <- input
	}
	close(to)
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

type CommandStore struct {
}

func (c *CommandStore) Lookup(name string) (Command, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		if err == exec.ErrNotFound {
			return nil, CommandMissingError{name}
		}

		return nil, fmt.Errorf("while looking up %q: %s", name, err)
	}

	return &NormalCommand{path}, nil
}

type NormalCommand struct {
	name string
}

func (c *NormalCommand) Run(out io.Writer, args ...string) error {
	if len(args) > 0 {
		args = args[1:]
	}

	cmd := exec.Command(c.name, args...)
	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

type CommandFinder interface {
	Lookup(name string) (Command, error)
}

type Command interface {
	Run(out io.Writer, args ...string) error
}

type CommandMissingError struct {
	name string
}

func (c CommandMissingError) Error() string {
	return fmt.Sprintf("%q: command not found", c.name)
}
