package devianter

import (
	"errors"
	"strconv"
	"strings"
)

// GRuser is a profile — a user's or a group's, as DeviantArt models both the
// same way. Owner.Group distinguishes them, and determines which of the
// ModuleData fields are populated: GroupAbout and GroupAdmins for a group, the
// embedded users for a person.
//
// The Page.Modules slice mirrors the site's own profile layout, so a caller
// looking for one piece of information has to search the slice for the module
// that carries it rather than reading a field directly.
type GRuser struct {
	ErrorDescription string
	Owner            struct {
		Group    bool `json:"isGroup"`
		Username string
	}
	Gruser struct {
		ID   int `json:"gruserId"`
		Page struct {
			Modules []struct {
				Name       string
				ModuleData struct {
					GroupAbout  GroupAbout
					GroupAdmins GroupAdmins
					users
				}
			}
		}
	}
	Extra struct {
		Tag   string `json:"gruserTagline"`
		Stats struct {
			Deviations, Watchers, Watching, Pageviews, CommentsMade, Favourites, Friends int
			FeedComments                                                                 int `json:"commentsReceivedProfile"`
		}
	} `json:"pageExtraData"`
}

// Gallery is a listing of deviations from a profile, returned by
// [Group.Gallery] and [Group.Favourites].
//
// Where the deviations land depends on the call. Results is the flat listing;
// folder-scoped requests instead nest them inside the Modules slice, under
// Folder for a gallery or Folders for the folder index itself.
type Gallery struct {
	Gruser struct {
		ID   int `json:"gruserId"`
		Page struct {
			Modules []struct {
				Name       string
				ModuleData struct {
					// Folders is the index of a profile's folders, each with a
					// representative thumbnail.
					Folders struct {
						HasMore bool
						Results []struct {
							Deviations int `json:"totalItemCount"`
							FolderId   int
							Size       int
							Name       string
							Thumb      Deviation
						}
					}

					// Folder is the contents of one folder.
					Folder struct {
						HasMore    bool
						Username   string
						Pages      int `json:"totalPageCount"`
						Deviations []Deviation
					} `json:"folderDeviations"`
				}
			}
		}
	}
	HasMore bool
	Results []Deviation
}

// Group is the entry point for everything scoped to one profile. Despite the
// name it addresses users as well as groups, since DeviantArt treats the two
// alike.
//
// Name is the profile's username and must be set; the methods return an error
// otherwise. Construct it directly:
//
//	g := devianter.Group{Name: "someuser"}
//	profile, apiErr, err := g.Get()
type Group struct {
	Name    string // required
	Content Gallery
}

// Get retrieves the profile itself — its about page, statistics, and, for a
// group, its admins. It works for both users and groups; inspect
// Owner.Group on the result to tell which was returned.
func (s Group) Get() (g GRuser, daError Error, err error) {
	if s.Name == "" {
		return g, daError, errors.New("missing Name field")
	}
	daError = ujson("dauserprofile/init/about?username="+s.Name, &g)

	return
}

// Favourites retrieves a page of the profile's favourites (its collections), 50
// at a time, zero-based.
//
// Set all to gather every folder's contents into one listing. Otherwise pass a
// positive folderid to read a single folder, or 0 for the profile's default
// favourites listing.
//
// folderid is optional; omitting it is the same as passing 0. Only the first
// value is used.
func (s Group) Favourites(page int, all bool, folderid ...int) (g Group, err Error) {
	var url strings.Builder

	fid := 0
	if len(folderid) > 0 {
		fid = folderid[0]
	}

	if fid > 0 || all {
		url.WriteString("dashared/gallection/contents")
		if all {
			url.WriteString("?all_folder=true")
		} else {
			url.WriteString("?folderid=")
			url.WriteString(strconv.Itoa(fid))
		}
		url.WriteString("&type=collection&")
	} else {
		url.WriteString("dauserprofile/init/favourites?deviations_")
	}

	url.WriteString("limit=50&username=")
	url.WriteString(s.Name)
	url.WriteString("&with_subfolders=true&offset=")
	url.WriteString(strconv.Itoa(page * 50))

	err = ujson(url.String(), &g.Content)
	return
}

// Gallery retrieves a page of the profile's gallery, 50 deviations at a time.
// Pass a positive folderid to read one folder, or 0 for the whole gallery.
//
// folderid is optional; omitting it is the same as passing 0. Only the first
// value is used.
//
// Note that page is interpreted differently by the two paths this takes: the
// whole-gallery listing is zero-based, while a folder listing is one-based.
func (s Group) Gallery(page int, folderid ...int) (g Group, daError Error, err error) {
	if s.Name == "" {
		return g, daError, errors.New("missing Name field")
	}

	fid := 0
	if len(folderid) > 0 {
		fid = folderid[0]
	}

	var url strings.Builder
	if fid > 0 {
		page--
		url.WriteString("dashared/gallection/contents?username=")
		url.WriteString(s.Name)
		url.WriteString("&folderid=")
		url.WriteString(strconv.Itoa(fid))
		url.WriteString("&offset=")
		url.WriteString(strconv.Itoa(page * 50))
		url.WriteString("&type=gallery&")
	} else {
		url.WriteString("dauserprofile/init/gallery?username=")
		url.WriteString(s.Name)
		url.WriteString("&page=")
		url.WriteString(strconv.Itoa(page))
		url.WriteString("&deviations_")
	}
	url.WriteString("limit=50")
	url.WriteString("&with_subfolders=false")

	daError = ujson(url.String(), &g.Content)
	return
}

// GroupAbout is a group's about page: when it was founded and its description.
type GroupAbout struct {
	FoundatedAt timeStamp `json:"foundationTs"`
	Description Text
}

// GroupAdmins lists a group's staff. TypeId encodes each member's role
// (founder, co-founder, contributor).
type GroupAdmins struct {
	Results []struct {
		TypeId int
		User   struct {
			Username string
		}
	}
}

// About is a person's profile information, all of it self-reported and any of
// it possibly empty.
type About struct {
	Country, Website, WebsiteLabel, Gender string
	// RegDate is how long the account has existed, in seconds — an age, not a
	// registration date, despite the name.
	RegDate     int64 `json:"deviantFor"`
	Description Text  `json:"textContent"`

	SocialLinks []struct {
		Value string
	}
	Interests []struct {
		Label, Value string
	}
}

// users is the person-specific half of a profile's module data, embedded into
// [GRuser] so its fields surface inline.
type users struct {
	About          About
	CoverDeviation struct {
		Deviation Deviation `json:"coverDeviation"`
	}
}
