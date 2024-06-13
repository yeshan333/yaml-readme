package function

import (
	"fmt"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

func IsLastServenDays(t *time.Time) bool {
	now := time.Now()
	diff := now.Sub(*t)
	return diff >= 0 && diff < 7*24*time.Hour
}

func GetLatestPost(items []*gofeed.Item) *gofeed.Item {
	length := len(items)
	if length == 0 {
		return nil
	}

	latest := items[0]
	for _, item := range items {
		if item.PublishedParsed == nil {
			continue
		}
		if item.PublishedParsed.After(*latest.PublishedParsed) {
			latest = item
		}
	}

	return latest
}

func GetFeedLatestPost(feedLink string, defaultContent string) (output string) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedLink)

	if err != nil {
		return fmt.Sprintf("[%s](%s)", defaultContent, defaultContent)

	}
	// get latest post
	latest := GetLatestPost(feed.Items)
	if latest == nil {
		return "feed parsed failed"
	}
	title := strings.ReplaceAll(latest.Title, "|", " ")
	link := latest.Link
	output = fmt.Sprintf(`[%s](%s)`, title, link)
	if IsLastServenDays(latest.PublishedParsed) {
		output += "![news](https://github.com/ChanceYu/front-end-rss/blob/master/assets/new.png?raw=true)"
	}
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

	if latest == nil {
		return "feed parsed failed"
	}
	publishTime := latest.PublishedParsed.Format(time.RFC3339)

	return publishTime
}
