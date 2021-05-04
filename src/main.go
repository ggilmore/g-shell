package main

import (
	"bufio"
	"flag"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

var command string

func main() {
	flag.StringVar(&command, "c", "", command)
	flag.Parse()

	var opts []ShellOption

	if isFlagPassed("c") {
		opts = append(opts, NonInteractive(command))
	}

	s := NewShell(opts...)
	s.Run()
}

func (s *shell) Run() {
	var wg sync.WaitGroup

	inputs := make(chan string)

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.runCommand(inputs)
	}()

	s.PrintPrompt()

	s.tokenize(inputs)
	wg.Wait()
}

func (s *shell) runCommand(inputs chan string) {
	errWriter := bufio.NewWriter(s.Stderr)

	writeErr := func(e error) {
		errWriter.WriteString(fmt.Sprintln(e.Error()))
		errWriter.Flush()

		s.PrintPrompt()
	}

	for input := range inputs {
		fields := strings.Fields(input)

		if len(fields) == 0 {
			s.PrintPrompt()
			continue
		}

		cmd := fields[0]

		c, err := s.cf.Lookup(cmd)
		if err != nil {
			writeErr(err)
			continue
		}

		rest := fields[1:]
		err = c.Run(s.Stdout, s.Stderr, rest...)
		if err != nil {
			if _, ok := err.(*exec.ExitError); !ok {
				writeErr(err)
			}
		}

		s.PrintPrompt()
	}
}

func (s *shell) tokenize(to chan string) {
	scanner := bufio.NewScanner(s.Stdin)

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
