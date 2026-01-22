package main

import (
	"context"
	"fmt"
)

func handlerResetDB(s *state, cmd command) error {
	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't delete users %v", err)
	}

	err = s.db.DeleteFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't delete users %v", err)
	}

	fmt.Println("Database reset successful.")
	return nil
}
