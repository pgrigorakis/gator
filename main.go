package main

import (
	"fmt"
	"github.com/pgrigorakis/gator/internal/config"
	"github.com/pgrigorakis/gator/cmd/commands"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	state :=    


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
