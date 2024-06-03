package devianter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
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

/* SEARCH */
type search struct {
	Total      int `json:"estTotal"`
	Pages      int // only for 'a' scope.
	Deviations []deviantion
}

func Search(query, page string, scope rune) (ss search) {
	var url strings.Builder

	// о5 построение ссылок.
	switch scope {
	case 'a':
		url.WriteString("dabrowse/search/all?q=")
	case 't':
		url.WriteString("dabrowse/networkbar/tag/deviations?tag=")
	default:
		log.Fatalln("Invalid type.\n- 'a' -- all;\n- 't' -- tag.")
	}

	url.WriteString(query)
	url.WriteString("&page=")
	url.WriteString(page)

	ujson(url.String(), &ss)

	// расчёт, сколько всего страниц по запросу. без токена 417 страниц - максимум
	for x := 0; x < int(math.Round(float64(ss.Total/25))); x++ {
		if x <= 417 {
			ss.Pages = x
		}
	}

	return
}

/* PUPPY aka DeviantArt API */
func puppy(data string) (string, error) {
	// получение или обновление токена
	update := func() (string, string, error) {
		var cookie string
		if cookie == "" {
			req := request("https://www.deviantart.com/_puppy")

			for _, content := range req.Cookies {
				cookie = content.Raw
			}
		}

		req := request("https://www.deviantart.com", cookie)
		if req.Status != 200 {
			return "", "", errors.New(req.Body)
		}

		return cookie, req.Body[strings.Index(req.Body, "window.__CSRF_TOKEN__ = '")+25 : strings.Index(req.Body, "window.__XHR_LOCAL__")-3], nil
	}

	// использование токена
	var (
		cookie, token string
	)
	if cookie == "" || token == "" {
		var e error
		cookie, token, e = update()
		if e != nil {
			return "", e
		}
	}

	body := request(
		fmt.Sprintf("https://www.deviantart.com/_puppy/%s&csrf_token=%s&da_minor_version=20230710", data, token),
		cookie,
	)

	// если код ответа не 200, возвращается ошибка
	if body.Status != 200 {
		return "", errors.New(body.Body)
	}

	return body.Body, nil
}
