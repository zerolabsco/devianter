package devianter

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// хрень для парсинга времени публикации
type timeStamp struct {
	time.Time
}

func (t *timeStamp) UnmarshalJSON(b []byte) (err error) {
	if b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	t.Time, err = time.Parse("2006-01-02T15:04:05-0700", string(b))
	return
}

// самая главная структура для поста
type Deviation struct {
	Title, Url, License string
	PublishedTime       timeStamp
	ID                  int `json:"deviationId"`

	NSFW bool `json:"isMature"`
	AI   bool `json:"isAiGenerated"`
	DD   bool `json:"isDailyDeviation"`

	Author struct {
		Username string
	}
	Stats struct {
		Favourites, Views, Downloads int
	}
	Media    Media
	Extended struct {
		Tags []struct {
			Name string
		}
		OriginalFile struct {
			Type     string
			Width    int
			Height   int
			Filesize int
		}
		DescriptionText Text
		RelatedContent  []struct {
			Deviations []Deviation
		}
	}
	TextContent Text
}

// её выпердыши
type Media struct {
	BaseUri string
	Name    string `json:"prettyName"`
	Token   []string
	Types   []struct {
		T    string
		H, W int
	}
}

type Text struct {
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
		Posted         timeStamp
		Replies, Likes int
	}

	IMG, Description string
}

// преобразование урла в правильный
func UrlFromMedia(m Media, thumb ...int) (urlParsed, wellFormattedFilename string) {
	var url strings.Builder

	subtractWidthHeight := func(to int, target ...*int) {
		for i, l := 0, len(target); i < l; i++ {
			for x := *target[i]; x > to; x -= to {
				*target[i] = x
			}
		}
	}

	for _, t := range m.Types {
		if t.T == "fullview" {
			url.WriteString(m.BaseUri)
			if l := len(m.BaseUri); l != 0 && (m.BaseUri[l-3:] != "gif" && t.W*t.H < 33177600) {
				if len(thumb) != 0 {
					subtractWidthHeight(thumb[0], &t.W, &t.H)
				}
				wellFormattedFilename = m.Name + m.BaseUri[l-4:]

				url.WriteString("/v1/fit/w_")
				url.WriteString(strconv.Itoa(t.W))
				url.WriteString(",h_")
				url.WriteString(strconv.Itoa(t.H))
				url.WriteString("/")
				url.WriteString(wellFormattedFilename)

			}
			if len(m.Token) > 0 {
				url.WriteString("?token=")
				url.WriteString(m.Token[0])
			}
		}
	}

	urlParsed = url.String()

	return
}

// для работы функции нужно ID поста и имя пользователя.
func GetDeviation(id string, user string) (st Post, err Error) {
	err = ujson(
		"dadeviation/init?deviationid="+id+"&username="+user+"&type=art&include_session=false&expand=deviation.related&preload=true",
		&st,
	)

	st.IMG, _ = UrlFromMedia(st.Deviation.Media)

	// базовая обработка описания
	txt := st.Deviation.TextContent.Html.Markup
	if len(txt) > 1 && txt[1] == '{' {
		var description struct {
			Blocks []struct {
				Text string
			}
		}

		if err := json.Unmarshal([]byte(txt), &description); err != nil {
			// Handle error appropriately
			try(err) // or log/return the error
		}
		for _, a := range description.Blocks {
			txt = a.Text
		}
	}
	st.Description = txt

	return
}
