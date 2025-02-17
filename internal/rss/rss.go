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
	// Make the HTTP GET request to the feedURL.
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "gator")

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// Populate the RSSFeed struct.
	xmlBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	rssFeed := &RSSFeed{}

	if err = xml.Unmarshal(xmlBytes, rssFeed); err != nil {
		return nil, err
	}

	// Decode escaped HTML entities.
	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Title)

	for _, rssItem := range rssFeed.Channel.Item {
		rssItem.Title = html.UnescapeString(rssItem.Title)
		rssItem.Description = html.UnescapeString(rssItem.Description)
	}

	return rssFeed, nil
}
