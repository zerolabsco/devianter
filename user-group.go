package devianter

import (
	"strconv"
	"strings"
)

// структура группы или пользователя
type groups struct {
	GroupAbout struct {
		FoundatedAt time `json:"foundationTs"`
		Description Text
	}
	GroupAdmins struct {
		Results []struct {
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
					groups
					users

					// группы
					Folders struct {
						Results []struct {
							FolderId int
							Name     string
						}
					}

					// галерея
					Folder struct {
						Username   string
						Pages      int `json:"totalPageCount"`
						Deviations []Deviation
					} `json:"folderDeviations"`
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

type Group struct {
	Name    string // обязательно заполнить
	Content GRuser
}

// подходит как группа, так и пользователь
func (s Group) GroupFunc() (g GRuser) {
	ujson("dauserprofile/init/about?username="+s.Name, &g)

	return
}

// гарелея пользователя или группы
func (s Group) Gallery(page int) (g Group) {
	var url strings.Builder
	url.WriteString("dauserprofile/init/gallery?username=")
	url.WriteString(s.Name)
	url.WriteString("&page=")
	url.WriteString(strconv.Itoa(page))
	url.WriteString("&deviations_limit=50&with_subfolders=false")

	ujson(url.String(), &g)
	return
}
