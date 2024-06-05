package devianter

import (
	"encoding/json"
	"strconv"
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
type deviantion struct {
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
			Deviations []deviantion
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
type Deviantion struct {
	Deviation deviantion
	Comments  struct {
		Total  int
		Cursor string
	}

	ParsedComments []struct {
		Author         string
		Posted         time
		Replies, Likes int
	}

	IMG, Desctiption string
}

// для работы функции нужно ID поста и имя пользователя.
func Deviation(id string, user string) Deviantion {
	var st Deviantion
	ujson(
		"dadeviation/init?deviationid="+id+"&username="+user+"&type=art&include_session=false&expand=deviation.related&preload=true",
		&st,
	)

	// преобразование урла в правильный
	for _, t := range st.Deviation.Media.Types {
		if m := st.Deviation.Media; t.T == "fullview" {
			if len(m.Token) > 0 {
				st.IMG = m.BaseUri + "?token="
			} else {
				st.IMG = m.BaseUri + "/v1/fill/w_" + strconv.Itoa(t.W) + ",h_" + strconv.Itoa(t.H) + "/" + id + "_" + user + ".gif" + "?token="
			}
			st.IMG += m.Token[0]
		}
	}

	// базовая обработка описания
	txt := st.Deviation.TextContent.Html.Markup
	if len(txt) > 0 && txt[1] == 125 {
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

	st.Desctiption = txt

	return st
}
