package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pgrigorakis/gator/internal/database"
	"github.com/pgrigorakis/gator/internal/rss"
)

func handlerRSS(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: %s <url>", cmd.name)
	}
	url := cmd.args[0]

	feed, err := rss.FetchFeed(context.Background(), url)
	if err != nil {
		return fmt.Errorf("couldn't fetch feed: %w", err)
	}

	fmt.Printf("Feed: %+v\n", feed)

	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("couldn't find user: %w", err)
	}

	if len(cmd.args) != 2 {
		return fmt.Errorf("usage: %s <feed name> <url>", cmd.name)
	}

	feedName := cmd.args[0]
	url := cmd.args[1]

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      feedName,
		Url:       url,
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("couldn't create feed: %w", err)
	}

	fmt.Printf("Feed '%v' added to user %v\n", feedName, user.Name)
	printFeed(feed)

	return nil
}

func printFeed(feed database.Feed) {
	fmt.Printf("* ID:            %s\n", feed.ID)
	fmt.Printf("* Created:       %v\n", feed.CreatedAt)
	fmt.Printf("* Updated:       %v\n", feed.UpdatedAt)
	fmt.Printf("* Name:          %s\n", feed.Name)
	fmt.Printf("* URL:           %s\n", feed.Url)
	fmt.Printf("* UserID:        %s\n", feed.UserID)
}
