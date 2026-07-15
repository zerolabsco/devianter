package devianter

import (
	"encoding/json"
	"html"
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
	//
	// Text is all it holds: a comment is a rich document, and its images,
	// emotes, mentions, and link targets are dropped in the flattening. Read
	// TextContent for those.
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
			cmmts.Thread[i].Comment = flattenMarkup(cmmts.Thread[i].TextContent.Html.Markup)
		}
	}

	return
}

// flattenMarkup renders a body of user-written markup as plain text, be it a
// comment or a deviation's description. Bodies are JSON inside JSON, and
// DeviantArt still serves all three formats it has used over the years:
//
//   - tiptap, current: {"version":1,"document":{"type":"doc","content":[...]}}
//   - Draft.js, legacy: {"blocks":[{"text":"..."}]}
//   - plain HTML, oldest, which passes through unchanged
//
// Block-level elements (paragraphs, headings) are joined with newlines, one per
// line, and a hard break inside one becomes a newline too. HTML entities in the
// text are decoded, so a body reads as &#8217; on the wire but an apostrophe
// here. Markup matching no known format passes through unchanged rather than
// being replaced by an empty string.
func flattenMarkup(m string) string {
	l := len(m)
	if l == 0 || m[0] != '{' || m[l-1] != '}' {
		return m
	}

	if text, ok := flattenTiptap(m); ok {
		return text
	}
	if text, ok := flattenDraftJS(m); ok {
		return text
	}
	return m
}

// tiptapNode is one node of a tiptap (ProseMirror) document tree.
//
// The document's "version" field is deliberately not modelled: DeviantArt sends
// it as a number on some bodies and a string on others, so any typed field for
// it fails to unmarshal on half of them.
type tiptapNode struct {
	Type    string       `json:"type"`
	Text    string       `json:"text"`
	Content []tiptapNode `json:"content"`
}

// flattenTiptap renders a tiptap document, reporting false if the markup is not
// one.
func flattenTiptap(m string) (string, bool) {
	var doc struct {
		Document tiptapNode `json:"document"`
	}
	if json.Unmarshal([]byte(m), &doc) != nil || doc.Document.Type != "doc" {
		return "", false
	}

	lines := make([]string, 0, len(doc.Document.Content))
	for _, block := range doc.Document.Content {
		var b strings.Builder
		writeTiptapText(block, &b)
		lines = append(lines, b.String())
	}

	return html.UnescapeString(strings.Join(lines, "\n")), true
}

// writeTiptapText collects the text of a node and everything nested inside it.
// Nodes carrying no text of their own — images, galleries, emotes — contribute
// nothing.
func writeTiptapText(n tiptapNode, b *strings.Builder) {
	switch n.Type {
	case "text":
		b.WriteString(n.Text)
		return
	case "hardBreak":
		b.WriteString("\n")
		return
	}
	for _, c := range n.Content {
		writeTiptapText(c, b)
	}
}

// flattenDraftJS renders a legacy Draft.js document, reporting false if the
// markup is not one.
func flattenDraftJS(m string) (string, bool) {
	var content struct {
		Blocks []struct {
			Text string
		}
	}
	if json.Unmarshal([]byte(m), &content) != nil || len(content.Blocks) == 0 {
		return "", false
	}

	lines := make([]string, 0, len(content.Blocks))
	for _, blk := range content.Blocks {
		lines = append(lines, blk.Text)
	}

	return html.UnescapeString(strings.Join(lines, "\n")), true
}
