package devianter

import (
	"encoding/json"
	"net/url"
	"strconv"
)

type Thread struct {
	Replies, Likes int
	ID             int `json:"commentId"`
	Parent         int `json:"parentId"`

	Posted timeStamp
	Author bool `json:"isAuthorHighlited"`

	Desctiption string
	Comment     string

	TextContent Text

	User struct {
		Username string
		Banned   bool `json:"isBanned"`
	}
}

type Comments struct {
	Cursor           string
	PrevOffset       int
	HasMore, HasLess bool

	Total  int
	Thread []Thread
}

// 1 - комментарии поста; 4 - комментарии на стене группы или пользователя
func GetComments(postid string, cursor string, page int, typ int) (cmmts Comments) {
	for x := 0; x <= page; x++ {
		ujson(
			"dashared/comments/thread?typeid="+strconv.Itoa(typ)+
				"&itemid="+postid+"&maxdepth=1000&order=newest"+
				"&limit=50&cursor="+url.QueryEscape(cursor),
			&cmmts,
		)

		cursor = cmmts.Cursor

		// парсинг json внутри json
		for i := 0; i < len(cmmts.Thread); i++ {
			m, l := cmmts.Thread[i].TextContent.Html.Markup, len(cmmts.Thread[i].TextContent.Html.Markup)
			cmmts.Thread[i].Comment = m

			// если начало строки {, а конец }, то срабатывает этот иф
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
