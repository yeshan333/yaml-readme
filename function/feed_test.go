package function

import (
	"testing"

	"github.com/mmcdole/gofeed"
)

func TestFeed(t *testing.T) {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL("https://github.com/SwiftOldDriver/iOS-Weekly/releases.atom")

	// some latest post in the last: https://wiki.eryajf.net/learning-weekly.xml
	latest := GetLatestPost(feed.Items)

	t.Log("Title: ", latest.Title)
	t.Log("Link: ", latest.Link)
	t.Log("Description: ", latest.Description)
	t.Log("Published: ", latest.Published)
	t.Log("Published: ", latest.PublishedParsed)
	t.Logf("[%s](%s)", latest.Title, latest.Link)
	// assert.Equal(t, "【转载】Go语言中的并发模型", latest.Title)
}
