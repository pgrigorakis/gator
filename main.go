package main

import (
	"fmt"

	"github.com/bohemian83/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

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
