package main

import _ "github.com/lib/pq"
import (
	"database/sql"
	"github.com/pgrigorakis/gator/internal/config"
	"github.com/pgrigorakis/gator/internal/database"
	"log"
	"os"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

func main() {

	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	dbQueries := database.New(db)

	programState := &state{db: dbQueries, cfg: &cfg}

	cmds := &commands{commandMap: make(map[string]func(*state, command) error)}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
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
