package function

import (
	"fmt"
	"time"

	"github.com/mmcdole/gofeed"
)

func GetLatestPost(items []*gofeed.Item) *gofeed.Item {
	length := len(items)
	if items[0].PublishedParsed.After(*items[length-1].PublishedParsed) {
		return items[0]
	}
	return items[length-1]
}

func GetFeedLatestPost(feedLink string, defaultContent string) (output string) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedLink)

	if err != nil {
		return fmt.Sprintf("[%s](%s)", defaultContent, defaultContent)

	}
	// get latest post
	latest := GetLatestPost(feed.Items)

	title := latest.Title
	link := latest.Link
	output = fmt.Sprintf("[%s](%s)", title, link)
	return output
}

func GetFeedLatestPostPublishedDate(feedLink string) (output string) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedLink)

	if err != nil {
		return ""
	}
	// get latest post
	latest := GetLatestPost(feed.Items)
	publishTime := latest.PublishedParsed.Format(time.RFC3339)

	return publishTime
}
