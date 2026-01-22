package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/pgrigorakis/gator/internal/database"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.name)
	}

	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("couldn't find user: %w", err)
	}
	err = s.cfg.SetUser(cmd.args[0])
	if err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Printf("User has been set to: %v\n", cmd.args[0])

	return nil

}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.name)
	}

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      cmd.args[0],
	})
	if err != nil {
		return fmt.Errorf("couldn't create user: %w", err)
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Printf("User created: %v\n", cmd.args[0])
	log.Printf("\nid: %v,\ncreated at: %v,\nupdated at: %v,\nname: %v\n", user.ID.String(), user.CreatedAt, user.UpdatedAt, user.Name)

	return nil

}

func handlerUsersList(s *state, cmd command) error {
	usersList, err := s.db.GetAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't list users: %w", err)
	}

	for i := range usersList {
		marker := ""
		if usersList[i].Name == s.cfg.CurrentUserName {
			marker = " (current)"
		}
		fmt.Printf("* %v%v\n", usersList[i].Name, marker)
	}

	return nil
}
