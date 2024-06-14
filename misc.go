package devianter

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
)

// функция для высера ошибки в stderr
func err(txt error) {
	if txt != nil {
		println(txt.Error())
	}
}

// сокращение для вызова щенка и парсинга жсона
func ujson(data string, output any) {
	input, e := puppy(data)
	err(e)

	eee := json.Unmarshal([]byte(input), output)
	err(eee)
}

/* REQUEST SECTION */
// структура для ответа сервера
type reqrt struct {
	Body    string
	Status  int
	Cookies []*http.Cookie
	Headers http.Header
}

// функция для совершения запроса
func request(uri string, other ...string) reqrt {
	var r reqrt

	// создаём новый запрос
	cli := &http.Client{}
	req, e := http.NewRequest("GET", uri, nil)
	err(e)

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:123.0) Gecko/20100101 Firefox/123.0.0")

	// куки и UA-шник
	if other != nil {
		for num, rng := range other {
			switch num {
			case 1:
				req.Header.Set("User-Agent", rng)
			case 0:
				req.Header.Set("Cookie", rng)
			}
		}
	}

	resp, e := cli.Do(req)
	err(e)
	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	err(e)

	// заполняем структуру
	r.Body = string(body)
	r.Cookies = resp.Cookies()
	r.Headers = resp.Header
	r.Status = resp.StatusCode

	return r
}

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

	return "", errors.New("User not exists")
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
	Total   int `json:"estTotal"`
	Pages   int // only for 'a' and 'g' scope.
	HasMore bool
	Results []Deviation `json:"deviations,results"`
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
			e = errors.New("Missing username (last argument)")
			return
		}
	default:
		log.Fatalln("Invalid type.\n- 'a' -- all;\n- 't' -- tag;\n- 'g' - gallery.")
	}

	url.WriteString(query)
	if scope != 'g' { // если область поиска не равна поиску по группам, то активируется этот код
		url.WriteString("&page=")
	} else { // иначе вместо страницы будет оффсет и страница умножится на 50
		url.WriteString("&offset=")
		page = 50 * page
	}
	url.WriteString(strconv.Itoa(page))

	ujson(url.String(), &ss)

	// расчёт, сколько всего страниц по запросу. без токена 417 страниц - максимум
	totalfloat := int(math.Round(float64(ss.Total / 25)))
	for x := 0; x < totalfloat; x++ {
		if x <= 417 {
			ss.Pages = x
		}
	}

	return
}

/* PUPPY aka DeviantArt API */
// получение или обновление токена
var cookie string
var token string

func UpdateCSRF() error {
	if cookie == "" {
		req := request("https://www.deviantart.com/_puppy")

		for _, content := range req.Cookies {
			cookie = content.Raw
		}
	}

	req := request("https://www.deviantart.com", cookie)
	if req.Status != 200 {
		return errors.New(req.Body)
	}
	token = req.Body[strings.Index(req.Body, "window.__CSRF_TOKEN__ = '")+25 : strings.Index(req.Body, "window.__XHR_LOCAL__")-3]

	return nil
}

func puppy(data string) (string, error) {
	var url strings.Builder
	url.WriteString("https://www.deviantart.com/_puppy/")
	url.WriteString(data)
	url.WriteString("&csrf_token=")
	url.WriteString(token)
	url.WriteString("&da_minor_version=20230710")

	body := request(url.String(), cookie)

	// если код ответа не 200, возвращается ошибка
	if body.Status != 200 {
		return "", errors.New(body.Body)
	}

	return body.Body, nil
}
