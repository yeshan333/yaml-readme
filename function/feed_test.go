package function

import (
	"testing"

	"github.com/mmcdole/gofeed"
)

func TestFeed(t *testing.T) {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL("https://quail.ink/op7418/feed/atom")

	// 获取最新的文章
	latest := feed.Items[0]
	t.Log("Title: ", latest.Title)
	t.Log("Link: ", latest.Link)
	t.Log("Description: ", latest.Description)
	t.Log("Published: ", latest.Published)
	t.Log("Published: ", latest.PublishedParsed)
	// assert.Equal(t, "【转载】Go语言中的并发模型", latest.Title)
}
