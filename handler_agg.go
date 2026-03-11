package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pgrigorakis/gator/internal/database"
	"github.com/pgrigorakis/gator/internal/rss"
)

func parseTime(pubDate string) (time.Time, error) {
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		// add others if needed
	}
	for _, format := range formats {
		t, err := time.Parse(format, pubDate)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse time: %s", pubDate)
}

func addPost(s *state, nextFeed database.Feed, item rss.RSSItem, ctx context.Context) error {
	parsedTime, parseErr := parseTime(item.PubDate)
	if parseErr != nil {
		log.Printf("could not parse publication time for %s: %v", item.Title, parseErr)
	}
	_, err := s.db.CreatePost(ctx, database.CreatePostParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Title:     item.Title,
		Url:       item.Link,
		Description: sql.NullString{
			String: item.Description,
			Valid:  item.Description != "",
		},
		PublishedAt: sql.NullTime{
			Time:  parsedTime,
			Valid: parseErr == nil,
		},
		FeedID: nextFeed.ID,
	})

	if err != nil {
		return err
	}

	return nil
}

func scrapeFeeds(s *state) {
	ctx := context.Background()
	nextFeed, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		log.Println("could not get next feed to fetch:", err)
		return
	}

	_, err = s.db.MarkFeedFetched(ctx, nextFeed.ID)
	if err != nil {
		log.Printf("could not mark feed %s as fetched: %v", nextFeed.Name, err)
		return
	}

	feed, err := rss.FetchFeed(ctx, nextFeed.Url)
	if err != nil {
		log.Printf("could not get feed %s to fetch: %v", nextFeed.Url, err)
		return
	}

	for _, item := range feed.Channel.Item {
		fmt.Printf("Found post: [%s] %s\n", feed.Channel.Description, item.Title)
		err = addPost(s, nextFeed, item, ctx)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				continue
			}
			log.Printf("could not create post %s: %v", item.Title, err)
		}
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
