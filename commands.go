package main

import (
	"fmt"
)

type command struct {
	Name string
	Args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func newCommands() *commands {
	return &commands{
		handlers: make(map[string]func(*state, command) error),
	}
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	handler, exist := c.handlers[cmd.Name]
	if !exist {
		return fmt.Errorf("Unknwon command: %v", cmd.Name)
	}
	return handler(s, cmd)
}
