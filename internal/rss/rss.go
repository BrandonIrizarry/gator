package rss

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
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

func (rssFeed RSSFeed) String() string {
	bodyBuffer := make([]string, 0, len(rssFeed.Channel.Item))

	for _, rssItem := range rssFeed.Channel.Item {
		rssItemStr := fmt.Sprintf("%v", rssItem)
		bodyBuffer = append(bodyBuffer, rssItemStr)
	}

	body := strings.Join(bodyBuffer, "\n")

	title := rssFeed.Channel.Title
	link := rssFeed.Channel.Link
	description := rssFeed.Channel.Description

	return fmt.Sprintf("Title: %s\nLink: %s\nDescription: %s\nItems: %v\n", title, link, description, body)
}

func (rssItem RSSItem) String() string {
	title := rssItem.Title
	link := rssItem.Link
	description := rssItem.Description
	pubDate := rssItem.PubDate

	return fmt.Sprintf("\tTitle: %s\n\tLink: %s\n\tDescription: %s\n\tPubDate: %s\n", title, link, description, pubDate)
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
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)

	for i := range rssFeed.Channel.Item {
		rssItem := &rssFeed.Channel.Item[i]

		rssItem.Title = html.UnescapeString(rssItem.Title)
		rssItem.Description = html.UnescapeString(rssItem.Description)
	}

	return rssFeed, nil
}
