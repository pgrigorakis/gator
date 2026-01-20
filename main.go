package main

import (
	"github.com/pgrigorakis/gator/internal/config"
	"log"
	"os"
)

type state struct {
	cfg *config.Config
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	programState := &state{cfg: &cfg}

	cmds := &commands{commandMap: make(map[string]func(*state, command) error)}
	cmds.register("login", handlerLogin)

	userArgs := os.Args
	if len(userArgs) < 2 {
		log.Fatal("error: no command entered")
	}

	cmd := command{name: userArgs[1], args: userArgs[2:]}

	err = cmds.run(programState, cmd)
	if err != nil {
		log.Fatal(err)
	}
}
