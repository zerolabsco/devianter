package devianter

import (
	"encoding/json"
	"errors"
	"io"
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
	for num, rng := range other {
		switch num {
		case 1:
			req.Header.Set("User-Agent", rng)
		case 0:
			req.Header.Set("Cookie", rng)
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
