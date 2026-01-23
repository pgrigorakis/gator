package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pgrigorakis/gator/internal/database"
)

func handlerFollowFeed(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: %s <url>", cmd.name)
	}
	url := cmd.args[0]

	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("couldn't find user: %w", err)
	}

	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("couldn't find feed with this url: %w", err)
	}

	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("couldn't create feed: %w", err)
	}

	fmt.Printf("* Feed Name:      %s\n", feedFollow.FeedName)
	fmt.Printf("* Current User:   %v\n", feedFollow.UserName)
	return nil

}

func handlerListFollowedFeeds(s *state, cmd command) error {
	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("couldn't find user: %w", err)
	}

	feeds, err := s.db.GetFeedFollowsByUserName(context.Background(), user.Name)
	if err != nil {
		return fmt.Errorf("couldn't find feed with this url: %w", err)
	}

	if len(feeds) == 0 {
		fmt.Println("User is following no feeds.")
		return nil
	}

	for _, feed := range feeds {
		fmt.Printf("* Name:          %s\n", feed.FeedName)
		fmt.Printf("* UserID:        %s\n", feed.UserName)
		fmt.Println("===================================")
	}
	return nil

}
