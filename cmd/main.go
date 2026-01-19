package main

import (
	"fmt"
	"github.com/pgrigorakis/gator/internal/config"
)

type state struct {
	cfg *config.Config
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	programState := &state{cfg: &cfg}
	commandsMap := make(map[string]func(*state, command) error)
	commands := &commands{commandsMap: &commandsMap}

	current_user := "potis"
	err = cfg.SetUser(current_user)
	if err != nil {
		fmt.Print(err)
	}

	cfg, err = config.Read()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	fmt.Printf("%v\n", cfg)
}
