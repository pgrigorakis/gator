package main

import (
	"fmt"
)

type command struct {
	name string
	args []string
}

type commands struct {
	commandMap map[string]func(*state, command) error
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("no username")
	}

	err := s.cfg.SetUser(cmd.args[0])
	if err != nil {
		return err
	}

	fmt.Printf("User has been set to: %v\n", cmd.args[0])

	return nil

}

func (c *commands) run(s *state, cmd command) error {
	handler, exists := c.commandMap[cmd.name]
	if !exists {
		return fmt.Errorf("unknown command: %v", cmd.name)
	}

	err := handler(s, cmd)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	return nil
}

func (c *commands) register(name string, f func(s *state, cmd command) error) {
	_, exists := c.commandMap[name]
	if !exists {
		c.commandMap[name] = f
	}
}
