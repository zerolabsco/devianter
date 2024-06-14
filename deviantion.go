package devianter

import (
	"encoding/json"
	"strconv"
	"strings"
	timelib "time"
)

// хрень для парсинга времени публикации
type time struct {
	timelib.Time
}

func (t *time) UnmarshalJSON(b []byte) (err error) {
	if b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	t.Time, err = timelib.Parse("2006-01-02T15:04:05-0700", string(b))
	return
}

// самая главная структура для поста
type Deviation struct {
	Title, Url, License string
	PublishedTime       time

	NSFW bool `json:"isMature"`
	AI   bool `json:"isAiGenerated"`
	DD   bool `json:"isDailyDeviation"`

	Author struct {
		Username string
	}
	Stats struct {
		Favourites, Views, Downloads int
	}
	Media    media
	Extended struct {
		Tags []struct {
			Name string
		}
		DescriptionText text
		RelatedContent  []struct {
			Deviations []Deviation
		}
	}
	TextContent text
}

// её выпердыши
type media struct {
	BaseUri string
	Token   []string
	Types   []struct {
		T    string
		H, W int
	}
}

type text struct {
	Excerpt string
	Html    struct {
		Markup, Type string
	}
}

// структура поста
type Post struct {
	Deviation Deviation
	Comments  struct {
		Total  int
		Cursor string
	}

	ParsedComments []struct {
		Author         string
		Posted         time
		Replies, Likes int
	}

	IMG, Description string
}

// преобразование урла в правильный
func UrlFromMedia(m media) string {
	var url strings.Builder
	for _, t := range m.Types {
		if t.T == "fullview" {
			url.WriteString(m.BaseUri)
			if m.BaseUri[len(m.BaseUri)-3:] != "gif" && t.W*t.H < 33177600 {
				url.WriteString("/v1/fill/w_")
				url.WriteString(strconv.Itoa(t.W))
				url.WriteString(",h_")
				url.WriteString(strconv.Itoa(t.H))
				url.WriteString("/")
				url.WriteString("image")
				url.WriteString(".gif")
			}
			if len(m.Token) > 0 {
				url.WriteString("?token=")
				url.WriteString(m.Token[0])
			}
		}
	}
	return url.String()
}

// для работы функции нужно ID поста и имя пользователя.
func DeviationFunc(id string, user string) Post {
	var st Post
	ujson(
		"dadeviation/init?deviationid="+id+"&username="+user+"&type=art&include_session=false&expand=deviation.related&preload=true",
		&st,
	)

	st.IMG = UrlFromMedia(st.Deviation.Media)

	// базовая обработка описания
	txt := st.Deviation.TextContent.Html.Markup
	if len(txt) > 0 && txt[1] == '{' {
		var description struct {
			Blocks []struct {
				Text string
			}
		}

		json.Unmarshal([]byte(txt), &description)
		for _, a := range description.Blocks {
			txt = a.Text
		}
	}

	st.Description = txt

	return st
}
