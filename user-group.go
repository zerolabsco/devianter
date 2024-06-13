package devianter

import (
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
					About struct {
						Country, Website, WebsiteLabel, Gender, Tagline string
						DeviantFor                                      int64
						SocialLinks                                     []struct {
							Value string
						}
						TextContent text
						Interests   []struct {
							Label, Value string
						}
					}
					CoverDeviation struct {
						Deviation Deviation `json:"coverDeviation"`
					}

					// группы
					GroupAbout struct {
						Tagline     string
						CreatinDate time `json:"foundationTs"`
						Description text
					}
					GroupAdmins struct {
						Results []struct {
							Username string
						}
					}
					Folders struct {
						Results []struct {
							FolderId int
							Name     string
						}
					}

					// галерея
					ModuleData struct {
						Folder struct {
							Username   string
							Pages      int `json:"totalPageCount"`
							Deviations []Deviation
						} `json:"folderDeviations"`
					}
				}
			}
		}
	}
	PageExtraData struct {
		GruserTagline string
		Stats         struct {
			Deviations, Watchers, Watching, Pageviews, CommentsMade, Favourites, Friends int
			FeedComments                                                                 int `json:"commentsReceivedProfile"`
		}
	}
}

type Group struct {
	Name    string // обязательно заполнить
	Content GRuser
}

// подходит как группа, так и пользователь
func (s Group) GroupFunc() (g Group) {
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
