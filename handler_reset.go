package main

import (
	"context"
	"fmt"
)

func handlerResetDB(s *state, cmd command) error {
	ctx := context.Background()

	err := s.db.DeleteUsers(ctx)
	if err != nil {
		return fmt.Errorf("couldn't delete users %v", err)
	}
	fmt.Println("Database reset successful.")
	return nil
}
