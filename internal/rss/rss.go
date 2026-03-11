package rss

import (
	"context"
	"encoding/xml"
	"html"
	"io"
	"net/http"
	"time"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequestWithContext(context.Background(), "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "gator")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	rssContent := &RSSFeed{}
	err = xml.Unmarshal(body, &rssContent)
	if err != nil {
		return nil, err
	}

	rssContent.Channel.Title = html.UnescapeString(rssContent.Channel.Title)
	rssContent.Channel.Description = html.UnescapeString(rssContent.Channel.Description)

	for i := range rssContent.Channel.Item {
		rssContent.Channel.Item[i].Title = html.UnescapeString(rssContent.Channel.Item[i].Title)
		rssContent.Channel.Item[i].Description = html.UnescapeString(rssContent.Channel.Item[i].Description)
	}

	return rssContent, nil
}
