package devianter

import (
	"errors"
	"math"
	"net/url"
	"strconv"
	"strings"
)

/* AVATARS AND EMOJIS */
// AEmedia fetches a user's avatar or a site emoji by name. t selects which:
// 'a' for an avatar, 'e' for an emoji.
//
// It returns the image data itself, not a URL. DeviantArt does not say which
// format a given name is stored in, so this tries .jpg, .png, and .gif in turn
// and returns the first that exists — up to three requests per call, and three
// for a name that does not exist.
//
// Passing any other t returns an error without making a request.
func AEmedia(name string, t rune) (string, error) {
	if len(name) < 2 {
		return "", errors.New("name must be specified")
	}
	var extensions = [3]string{
		".jpg",
		".png",
		".gif",
	}
	name = strings.ToLower(name)

	// Avatars and emoji are sharded into directories by the leading characters of
	// the name; avatars additionally normalise dashes to underscores first.
	var b strings.Builder
	switch t {
	case 'a':
		b.WriteString("https://a.deviantart.net/avatars-big/")
		name_without_dashes := strings.ReplaceAll(name, "-", "_")
		b.WriteString(name_without_dashes[:1])
		b.WriteString("/")
		b.WriteString(name_without_dashes[1:2])
		b.WriteString("/")
	case 'e':
		b.WriteString("https://e.deviantart.net/emoticons/")
		b.WriteString(name[:1])
		b.WriteString("/")
	default:
		return "", errors.New("invalid type: want 'a' (avatar) or 'e' (emoji)")
	}
	b.WriteString(name)

	// Probe each extension; the first 200 is the real format.
	for x := 0; x < len(extensions); x++ {
		req := request(b.String() + extensions[x])
		if req.Status == 200 {
			return req.Body, nil
		}
	}

	return "", errors.New("user not exists")
}

/* DAILY DEVIATIONS */
// DailyDeviations is the staff-curated front page selection. The picks are
// grouped into Strips, each a titled row as the site presents it; Deviations is
// the ungrouped listing.
type DailyDeviations struct {
	HasMore bool
	Strips  []struct {
		Codename, Title string
		TitleType       string
		Deviations      []Deviation
	}
	Deviations []Deviation
}

// GetDailyDeviations retrieves a page of the daily deviation selection. Pages
// are zero-based; check the returned HasMore before asking for the next.
func GetDailyDeviations(page int) (dd DailyDeviations, err Error) {
	err = ujson("dabrowse/networkbar/rfy/deviations?page="+strconv.Itoa(page), &dd)
	return
}

/* SEARCH */
// Search is a page of search results. Read the matches from Results, which
// [PerformSearch] populates whichever field the endpoint used.
//
// Total is DeviantArt's own estimate and is approximate. Pages is derived from
// it and capped at 417, the depth a guest session can reach before the API stops
// paginating.
type Search struct {
	Total   int `json:"estTotal"`
	Pages   int // only for 'a' and 'g' scope.
	HasMore bool
	Results []Deviation `json:"deviations"`
	// ResultsGalleryTemp receives the results of gallery and collection searches,
	// which return them under a different key. PerformSearch copies it into
	// Results; callers should not need this field.
	ResultsGalleryTemp []Deviation `json:"results"`
}

// PerformSearch searches DeviantArt. scope selects what is being searched:
//
//	'a' — everything, by title and description
//	't' — by tag
//	'g' — within one user's or group's gallery
//	'f' — within one user's or group's collections (favourites)
//
// Scopes 'g' and 'f' search a particular account, so they require the username
// as the final argument and return an error without it. The other two ignore it.
//
// Pages are zero-based. A guest session cannot page beyond roughly 417 pages
// deep regardless of how many results Total claims.
//
// Passing any other scope returns an error without making a request.
func PerformSearch(query string, page int, scope rune, user ...string) (ss Search, daError Error, err error) {
	var buildurl strings.Builder

	switch scope {
	case 'a':
		buildurl.WriteString("dabrowse/search/all?q=")
	case 't':
		buildurl.WriteString("dabrowse/networkbar/tag/deviations?tag=")
	case 'g', 'f':
		if user == nil {
			err = errors.New("missing username (last argument)")
			return
		}

		buildurl.WriteString("dashared/gallection/search?username=")
		buildurl.WriteString(user[0])
		buildurl.WriteString("&type=")
		if scope == 'g' {
			buildurl.WriteString("gallery")
		} else {
			buildurl.WriteString("collection")
		}
		buildurl.WriteString("&order=most-recent&init=true&limit=50&q=")
	default:
		err = errors.New("invalid scope: want 'a' (all), 't' (tag), 'g' (gallery) or 'f' (favourites)")
		return
	}

	buildurl.WriteString(url.QueryEscape(query))
	// Gallery search paginates by item offset rather than page number.
	if scope != 'g' {
		buildurl.WriteString("&page=")
	} else {
		buildurl.WriteString("&offset=")
		page = 50 * page
	}
	buildurl.WriteString(strconv.Itoa(page))

	daError = ujson(buildurl.String(), &ss)

	if ss.Results == nil {
		ss.Results = ss.ResultsGalleryTemp
	}

	// Derive the page count from the result estimate, clamped to the 417 pages a
	// guest session can actually reach.
	totalfloat := int(math.Round(float64(ss.Total / 25)))
	for x := 0; x < totalfloat; x++ {
		if x <= 417 {
			ss.Pages = x
		}
	}

	return
}
