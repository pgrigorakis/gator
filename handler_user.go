package main

import (
	"context"
	"errors"
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

	ctx := context.Background()

	_, err := s.db.GetUser(ctx, cmd.args[0])
	if err != nil {
		return errors.New("no such user exists")
	}
	err = s.cfg.SetUser(cmd.args[0])
	if err != nil {
		return err
	}

	fmt.Printf("User has been set to: %v\n", cmd.args[0])

	return nil

}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.name)
	}
	ctx := context.Background()

	now := time.Now()

	_, err := s.db.GetUser(ctx, cmd.args[0])
	if err == nil {
		return errors.New("user exists already")
	} else {
		registeredUser, err := s.db.CreateUser(ctx, database.CreateUserParams{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
			Name:      cmd.args[0],
		})
		if err != nil {
			return err
		}

		err = s.cfg.SetUser(cmd.args[0])
		if err != nil {
			return err
		}

		fmt.Printf("User has been set to: %v\n", cmd.args[0])
		log.Printf("\nid: %v,\ncreated at: %v,\nupdated at: %v,\nname: %v\n", registeredUser.ID.String(), registeredUser.CreatedAt, registeredUser.UpdatedAt, registeredUser.Name)
	}

	return nil

}
