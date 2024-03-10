package function

import (
	"fmt"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

func IsThisWeek(t *time.Time) bool {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := now.AddDate(0, 0, -weekday+1)
	monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
	diff := t.Sub(monday)
	return diff >= 0 && diff < 7*24*time.Hour
}

func GetLatestPost(items []*gofeed.Item) *gofeed.Item {
	length := len(items)
	if length == 0 {
		return nil
	}
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
	if latest == nil {
		return "no content"
	}
	title := strings.ReplaceAll(latest.Title, "|", " ")
	link := latest.Link
	output = fmt.Sprintf("[%s](%s)", title, link)
	if IsThisWeek(latest.PublishedParsed) {
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
		return "no content"
	}
	publishTime := latest.PublishedParsed.Format(time.RFC3339)

	return publishTime
}
