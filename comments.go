package devianter

import (
	"encoding/json"
	"net/url"
	"strconv"
)

// Thread is a single comment, despite the name. Replies are not nested inside
// it: a thread arrives flattened into [Comments].Thread, and the shape is
// recovered through Parent, which holds the ID of the comment being replied to
// and is 0 for a top-level comment.
type Thread struct {
	Replies, Likes int
	ID             int `json:"commentId"`
	Parent         int `json:"parentId"`

	Posted timeStamp
	// Author reports whether the commenter is the author of the deviation being
	// commented on.
	Author bool `json:"isAuthorHighlited"`

	Desctiption string

	// Comment is the comment's plain text, which [GetComments] extracts from
	// TextContent. Prefer it; TextContent is the unprocessed original.
	Comment string

	TextContent Text

	User struct {
		Username string
		Banned   bool `json:"isBanned"`
	}
}

// Comments is one page of comments. Thread holds the comments themselves,
// flattened rather than nested; Total counts every comment on the item, not just
// this page. Cursor resumes from the end of this page and HasMore reports
// whether anything remains.
type Comments struct {
	Cursor           string
	PrevOffset       int
	HasMore, HasLess bool

	Total  int
	Thread []Thread
}

// GetComments retrieves comments on an item, 50 per page, with each comment's
// plain text extracted into [Thread].Comment.
//
// typ selects what postid refers to: 1 for comments on a deviation, 4 for those
// on a user's or group's profile wall. cursor resumes from a previous call's
// [Comments].Cursor; pass an empty string to start from the newest comment.
//
// page is an offset from cursor rather than an absolute page number, and it is
// walked one request at a time: page 5 costs six round-trips and returns only
// the sixth page. Paginating by feeding each result's Cursor back in with page 0
// costs one request per page, and is the cheaper way to walk a long thread.
func GetComments(postid string, cursor string, page int, typ int) (cmmts Comments, err Error) {
	for x := 0; x <= page; x++ {
		err = ujson(
			"dashared/comments/thread?typeid="+strconv.Itoa(typ)+
				"&itemid="+postid+"&maxdepth=1000&order=newest"+
				"&limit=50&cursor="+url.QueryEscape(cursor),
			&cmmts,
		)

		cursor = cmmts.Cursor

		// Comment bodies are JSON inside JSON: newer ones are a Draft.js document
		// encoded into the markup string, older ones are plain HTML. Flatten the
		// former to plain text and pass the latter through.
		for i := 0; i < len(cmmts.Thread); i++ {
			m, l := cmmts.Thread[i].TextContent.Html.Markup, len(cmmts.Thread[i].TextContent.Html.Markup)
			cmmts.Thread[i].Comment = m

			if m[0] == '{' && m[l-1] == '}' {
				var content struct {
					Blocks []struct {
						Text string
					}
				}

				e := json.Unmarshal([]byte(m), &content)
				try(e)

				for _, a := range content.Blocks {
					cmmts.Thread[i].Comment = a.Text
				}
			}
		}
	}

	return
}
