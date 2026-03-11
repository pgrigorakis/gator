package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pgrigorakis/gator/internal/rss"
)

func scrapeFeeds(s *state) {

	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		log.Println("could not get next feed to fetch:", err)
		return
	}

	_, err = s.db.MarkFeedFetched(context.Background(), nextFeed.ID)
	if err != nil {
		log.Printf("could not mark feed %s as fetched: %v", nextFeed.Name, err)
		return
	}

	feed, err := rss.FetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		log.Printf("could not get feed %s to fetch: %v", nextFeed.Url, err)
		return
	}

	for _, item := range feed.Channel.Item {
		fmt.Printf("Found post: [%s] %s\n", feed.Channel.Description, item.Title)
	}

	log.Printf("Feed %s collected, %v posts found", nextFeed.Name, len(feed.Channel.Item))
}

func handlerRSS(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: %s <interval e.g. 1s, 1m30s, 1h30m>", cmd.name)
	}
	interval := cmd.args[0]

	timeBetweenRequests, err := time.ParseDuration(interval)
	if err != nil {
		return fmt.Errorf("could not parse internal: %w\n", err)
	}

	ticker := time.NewTicker(timeBetweenRequests)

	fmt.Printf("Collecting feeds every %v...\n", interval)

	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}
