package main

import (
	"context"
	"fmt"
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
