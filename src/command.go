package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

type CommandFinder interface {
	Lookup(name string) (Command, error)
}

type Command interface {
	Run(stdout, stderr io.Writer, args ...string) error
}

type CommandStore struct {
	builtins map[string]Command
}

func (cs *CommandStore) Lookup(name string) (Command, error) {
	command, ok := cs.builtins[name]
	if ok {
		return command, nil
	}

	_, err := exec.LookPath(name)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, &CommandNotFoundError{
				name: name,
			}
		}

		return nil, fmt.Errorf("while looking up %q: %s", name, err)
	}

	return NewNormalCommand(name), nil
}

func (cs *CommandStore) LoadBuiltin(name string, c Command) {
	if cs.builtins == nil {
		cs.builtins = make(map[string]Command)
	}

	cs.builtins[name] = c
}

type NormalCommand struct {
	name string

	signals chan os.Signal
	done    chan struct{}
}

func NewNormalCommand(name string) *NormalCommand {
	return &NormalCommand{
		name: name,

		signals: make(chan os.Signal),
		done:    make(chan struct{}),
	}
}

func (c *NormalCommand) Run(stdin io.Writer, stdout io.Writer, args ...string) error {
	cmd := exec.Command(c.name, args...)

	cmd.Stdout = stdin
	cmd.Stderr = stdout

	signal.Notify(c.signals, syscall.SIGINT)

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("when starting %q: %w", c.name, err)
	}

	defer func() {
		c.done <- struct{}{}
	}()

	go func() {
		for {
			select {
			case s := <-c.signals:
				cmd.Process.Signal(s)
			case <-c.done:
				return
			}
		}
	}()

	return cmd.Wait()
}

type CommandNotFoundError struct {
	name string
	Err  error
}

func (e *CommandNotFoundError) Error() string {
	message := fmt.Sprintf("%q: command not found", e.name)
	if e.Err != nil {
		return fmt.Sprintf("%s - additional context: %s", message, e.Err.Error())
	}

	return message

}

func (e *CommandNotFoundError) Unwrap() error {
	return e.Err
}
