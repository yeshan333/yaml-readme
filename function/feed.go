package function

import (
	"fmt"

	"github.com/mmcdole/gofeed"
)

func GetLatestFeedPost(feedLink string) (output string) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedLink)

	if err != nil {
		return ""

	}
	// get latest post
	latest := feed.Items[0]

	title := latest.Title
	link := latest.Link
	publishTime := latest.Published
	output = fmt.Sprintf("%s: [%s](%s)", publishTime, title, link)
	// t.Log("Title: ", latest.Title)
	// t.Log("Link: ", latest.Link)
	// t.Log("Description: ", latest.Description)
	// t.Log("Published: ", latest.Published)
	// t.Log("Published: ", latest.PublishedParsed)
	return output
}
