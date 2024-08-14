package devianter

import (
	"errors"
	"strconv"
	"strings"
)

// структура группы или пользователя
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

type Gallery struct {
	Gruser struct {
		ID   int `json:"gruserId"`
		Page struct {
			Modules []struct {
				Name       string
				ModuleData struct {
					// группы
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

					// галерея
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

type Group struct {
	Name    string // обязательно заполнить
	Content Gallery
}

// подходит как группа, так и пользователь
func (s Group) Get() (g GRuser, err error) {
	if s.Name == "" {
		return g, errors.New("missing Name field")
	}
	ujson("dauserprofile/init/about?username="+s.Name, &g)

	return
}

func (s Group) Favourites(page int, all bool, folderid ...int) (g Group) {
	var url strings.Builder

	if fid := folderid[0]; fid > 0 || all {
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

	ujson(url.String(), &g.Content)
	return
}

// гарелея пользователя или группы
func (s Group) Gallery(page int, folderid ...int) (g Group, err error) {
	if s.Name == "" {
		return g, errors.New("missing Name field")
	}

	var url strings.Builder
	if folderid[0] > 0 {
		page--
		url.WriteString("dashared/gallection/contents?username=")
		url.WriteString(s.Name)
		url.WriteString("&folderid=")
		url.WriteString(strconv.Itoa(folderid[0]))
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

	ujson(url.String(), &g.Content)
	return
}

type GroupAbout struct {
	FoundatedAt timeStamp `json:"foundationTs"`
	Description Text
}
type GroupAdmins struct {
	Results []struct {
		TypeId int
		User   struct {
			Username string
		}
	}
}

type About struct {
	Country, Website, WebsiteLabel, Gender string
	RegDate                                int64 `json:"deviantFor"`
	Description                            Text  `json:"textContent"`

	SocialLinks []struct {
		Value string
	}
	Interests []struct {
		Label, Value string
	}
}

type users struct {
	About          About
	CoverDeviation struct {
		Deviation Deviation `json:"coverDeviation"`
	}
}
