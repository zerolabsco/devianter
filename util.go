package devianter

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// try prints a non-nil error to stderr and swallows it. It is how this package
// reports problems it does not propagate, such as a response that parsed only
// partially.
func try(txt error) {
	if txt != nil {
		println(txt.Error())
	}
}

// ujson fetches a _puppy endpoint and unmarshals the response into output.
// data is the path and query string after the endpoint root, without a leading
// slash and without the csrf_token parameter, which puppy appends.
//
// A malformed response is reported through try and leaves output partially
// populated, so a returned Error with an empty Reason does not by itself
// guarantee that output is complete.
func ujson(data string, output any) Error {
	input, err := puppy(data)
	if err == nil {
		try(json.Unmarshal([]byte(input), output))
	}
	return APIError(err)
}

// Error is a failed API call. It is a struct rather than an error interface, so
// a zero value means success: test Reason for emptiness rather than comparing
// against nil.
//
// For errors DeviantArt itself reports, Reason and Error hold its machine and
// human readable descriptions. For anything else (a transport failure, or a
// CloudFront block page) Reason is "request_failed" and Error carries the
// underlying message.
type Error struct {
	Reason string `json:"error"`
	Error  string `json:"errorDescription"`
	RAW    []byte `json:"-"`
}

// APIError converts an error from the request layer into an [Error], decoding
// DeviantArt's JSON error body when that is what it is. A nil input yields the
// zero Error, which signals success.
func APIError(inputError error) (err Error) {
	if inputError != nil {
		err.RAW = []byte(inputError.Error())
		// DA's API errors are JSON. Anything else (CDN block pages, transport
		// failures) is surfaced as-is rather than spamming a JSON parse error —
		// this is what used to print `invalid character '<'` on every page.
		if json.Unmarshal(err.RAW, &err) != nil {
			err.Reason = "request_failed"
			err.Error = inputError.Error()
		}
	}
	return
}

/* REQUEST SECTION */
// reqrt is a completed HTTP response, flattened into the pieces this package
// needs. On a transport failure Err is set and every other field is zero.
type reqrt struct {
	Body    string
	Status  int
	Cookies []*http.Cookie
	Headers http.Header
	// Err is set when the request never completed (transport error). Status is 0.
	Err error
}

// UserAgent overrides the browser User-Agent this package sends by default.
// Setting it to something that identifies your client is polite, but DeviantArt
// is more likely to serve a block page to a non-browser agent.
var UserAgent string

// Timeout bounds a single request end-to-end (dial, response, body read).
// Without it, a hung connection blocks its caller forever.
var Timeout = 30 * time.Second

// request performs a GET and never panics or returns a partial response without
// saying so: any failure is reported in reqrt.Err. An optional second argument
// supplies the Cookie header.
func request(uri string, other ...string) reqrt {
	var r reqrt

	// Transport is deliberately left nil so http.DefaultTransport applies: that
	// keeps HTTPS_PROXY support and lets callers wrap it (e.g. to rate-limit).
	cli := &http.Client{Timeout: Timeout}
	req, e := http.NewRequest("GET", uri, nil)
	if e != nil {
		try(e)
		r.Err = e
		return r
	}

	// Impersonate a browser by default: the endpoints are the web frontend's own,
	// and an unfamiliar agent draws a block page.
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:123.0) Gecko/20100101 Firefox/123.0.0")

	if UserAgent != "" {
		req.Header.Set("User-Agent", UserAgent)
	}
	if len(other) != 0 {
		req.Header.Set("Cookie", other[0])
	}

	resp, e := cli.Do(req)
	if e != nil {
		// resp is nil on error: returning here avoids dereferencing it, which
		// used to panic and (from UpdateCSRF's goroutine) kill the process.
		try(e)
		r.Err = e
		return r
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			try(err)
		}
	}()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		try(e)
		r.Err = e
	}

	r.Body = string(body)
	r.Cookies = resp.Cookies()
	r.Headers = resp.Header
	r.Status = resp.StatusCode

	return r
}

// looksLikeJSON reports whether a response is actually JSON, so an HTML page from
// a CDN/edge never reaches json.Unmarshal.
func looksLikeJSON(r reqrt) bool {
	if ct := r.Headers.Get("Content-Type"); ct != "" && !strings.Contains(ct, "json") {
		return false
	}
	b := strings.TrimSpace(r.Body)
	return len(b) > 0 && (b[0] == '{' || b[0] == '[')
}

// describe renders a failed response as a readable message, instead of the opaque
// `invalid character '<'` you get from json.Unmarshal on an HTML error page.
func describe(r reqrt) string {
	body := strings.TrimSpace(r.Body)
	if looksLikeJSON(r) {
		return body // DA's own JSON error; callers unmarshal it into Error
	}

	msg := "devianter: HTTP " + strconv.Itoa(r.Status) + " non-JSON response from DeviantArt"
	if strings.Contains(body, "Generated by cloudfront") || strings.Contains(body, "Request blocked") {
		msg += ": blocked by CloudFront/WAF — this egress IP is likely banned"
	}
	if len(body) > 200 {
		body = body[:200] + "..."
	}
	return msg + " — " + body
}

/* PUPPY aka DeviantArt API */
// The guest session: a cookie from the _puppy endpoint and a CSRF token scraped
// from the homepage. UpdateCSRF populates both; puppy sends them on every call.
//
// These are package-level and unsynchronised, so a program that calls UpdateCSRF
// concurrently with any other function of this package races on them.
var cookie string
var token string

const (
	csrfPrefix = "window.__CSRF_TOKEN__ = '"
	xhrMarker  = "window.__XHR_LOCAL__"
)

// UpdateCSRF establishes the guest session that every other call in this package
// depends on, and must be called before them. It fetches a session cookie (only
// on the first call; later calls reuse it) and scrapes a fresh CSRF token from
// the DeviantArt homepage.
//
// Tokens expire, so a long-running program should call this again when requests
// begin to fail. It is not safe to call concurrently with other functions of
// this package.
//
// An error means the session was not established: the homepage was blocked,
// served a challenge, or changed its markup such that the token is no longer
// where this package looks for it.
func UpdateCSRF() error {
	if cookie == "" {
		req := request("https://www.deviantart.com/_puppy")

		for _, content := range req.Cookies {
			cookie = content.Raw
		}
	}

	req := request("https://www.deviantart.com", cookie)
	if req.Err != nil {
		return req.Err
	}
	if req.Status != 200 {
		return errors.New(describe(req))
	}

	// Bounds-check the markers. On a block/challenge page they are absent, and the
	// old arithmetic sliced Body[24:-4] — a panic that killed the whole process.
	start, end := strings.Index(req.Body, csrfPrefix), strings.Index(req.Body, xhrMarker)
	if start < 0 || end < 0 {
		return errors.New("devianter: CSRF token not found in homepage (blocked, challenged, or markup changed)")
	}
	start += len(csrfPrefix)
	end -= 3
	if end <= start || end > len(req.Body) {
		return errors.New("devianter: CSRF token markers out of order (markup changed)")
	}
	token = req.Body[start:end]

	return nil
}

// puppy calls a _puppy endpoint with the guest session applied and returns the
// raw JSON body. data is a path and query string; the CSRF token and API version
// are appended to it, so it must already end in a parameter (callers conclude
// theirs with a trailing "&" or a final value).
//
// It returns an error for a transport failure, a non-200 status, or a 200 whose
// body is not JSON, which is how a CDN block page arrives.
func puppy(data string) (string, error) {
	var url strings.Builder
	url.WriteString("https://www.deviantart.com/_puppy/")
	url.WriteString(data)
	url.WriteString("&csrf_token=")
	url.WriteString(token)
	url.WriteString("&da_minor_version=20230710")

	body := request(url.String(), cookie)
	if body.Err != nil {
		return "", body.Err
	}

	if body.Status != 200 {
		return "", errors.New(describe(body))
	}

	// A 200 that isn't JSON means an edge/CDN page slipped through.
	if !looksLikeJSON(body) {
		return "", errors.New(describe(body))
	}

	return body.Body, nil
}
