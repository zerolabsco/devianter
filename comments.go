package devianter

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
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

		for i := 0; i < len(cmmts.Thread); i++ {
			cmmts.Thread[i].Comment = flattenComment(cmmts.Thread[i].TextContent.Html.Markup)
		}
	}

	return
}

// flattenComment renders a body of user-written markup as plain text, be it a
// comment or a deviation's description. Bodies are JSON inside JSON: newer ones
// are a Draft.js document encoded into the markup string, older ones are plain
// HTML, which passes through unchanged. Markup that does not parse, and empty
// markup, also pass through.
//
// A Draft.js document is a list of blocks, which are block-level elements
// (paragraphs, list items); they are joined with newlines, one block per line.
func flattenComment(m string) string {
	l := len(m)
	if l == 0 || m[0] != '{' || m[l-1] != '}' {
		return m
	}

	var content struct {
		Blocks []struct {
			Text string
		}
	}

	e := json.Unmarshal([]byte(m), &content)
	try(e)

	if len(content.Blocks) == 0 {
		return m
	}

	var b strings.Builder
	for i, a := range content.Blocks {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(a.Text)
	}
	return b.String()
}
