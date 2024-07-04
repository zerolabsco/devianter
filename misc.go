package devianter

import (
	"errors"
	"log"
	"math"
	u "net/url"
	"strconv"
	"strings"
)

/* AVATARS AND EMOJIS */
func AEmedia(name string, t rune) (string, error) {
	// список всех возможных расширений
	var extensions = [3]string{
		".jpg",
		".png",
		".gif",
	}
	// надо
	name = strings.ToLower(name)

	// построение ссылок. билдер потому что он быстрее обычного сложения строк.
	var b strings.Builder
	switch t {
	case 'a':
		b.WriteString("https://a.deviantart.net/avatars-big/")
		b.WriteString(name[:1])
		b.WriteString("/")
		b.WriteString(name[1:2])
		b.WriteString("/")
	case 'e':
		b.WriteString("https://e.deviantart.net/emoticons/")
		b.WriteString(name[:1])
		b.WriteString("/")
	default:
		log.Fatalln("Invalid type.\n- 'a' -- avatar;\n- 'e' -- emoji.")
	}
	b.WriteString(name)

	// проверка ссылки на доступность
	for x := 0; x < len(extensions); x++ {
		req := request(b.String() + extensions[x])
		if req.Status == 200 {
			return req.Body, nil
		}
	}

	return "", errors.New("user not exists")
}

/* DAILY DEVIATIONS */
type DailyDeviations struct {
	HasMore bool
	Strips  []struct {
		Codename, Title string
		TitleType       string
		Deviations      []Deviation
	}
	Deviations []Deviation
}

func DailyDeviationsFunc(page int) (dd DailyDeviations) {
	ujson("dabrowse/networkbar/rfy/deviations?page="+strconv.Itoa(page), &dd)
	return
}

/* SEARCH */
type Search struct {
	Total              int `json:"estTotal"`
	Pages              int // only for 'a' and 'g' scope.
	HasMore            bool
	Results            []Deviation `json:"deviations"`
	ResultsGalleryTemp []Deviation `json:"results"`
}

func SearchFunc(query string, page int, scope rune, user ...string) (ss Search, e error) {
	var url strings.Builder
	e = nil

	// о5 построение ссылок.
	switch scope {
	case 'a': // поиск артов по названию
		url.WriteString("dabrowse/search/all?q=")
	case 't': // поиск артов по тегам
		url.WriteString("dabrowse/networkbar/tag/deviations?tag=")
	case 'g': // поиск артов пользователя или группы
		if user != nil {
			url.WriteString("dashared/gallection/search?username=")
			for _, a := range user {
				url.WriteString(a)
			}
			url.WriteString("&type=gallery&order=most-recent&init=true&limit=50&q=")
		} else {
			e = errors.New("missing username (last argument)")
			return
		}
	default:
		log.Fatalln("Invalid type.\n- 'a' -- all;\n- 't' -- tag;\n- 'g' - gallery.")
	}

	url.WriteString(u.QueryEscape(query))
	if scope != 'g' { // если область поиска не равна поиску по группам, то активируется этот код
		url.WriteString("&page=")
	} else { // иначе вместо страницы будет оффсет и страница умножится на 50
		url.WriteString("&offset=")
		page = 50 * page
	}
	url.WriteString(strconv.Itoa(page))

	ujson(url.String(), &ss)

	if scope == 'g' {
		ss.Results = ss.ResultsGalleryTemp
	}

	// расчёт, сколько всего страниц по запросу. без токена 417 страниц - максимум
	totalfloat := int(math.Round(float64(ss.Total / 25)))
	for x := 0; x < totalfloat; x++ {
		if x <= 417 {
			ss.Pages = x
		}
	}

	return
}
