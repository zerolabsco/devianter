package devianter

import (
	"strconv"
	"strings"
)

// структура группы или пользователя
type Group struct {
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
						Deviation deviantion `json:"coverDeviation"`
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
							Deviations []deviantion
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

func UGroup(name string) (g Group) {
	ujson("dauserprofile/init/about?username="+name, &g)

	return
}

// гарелея пользователя или группы
func Gallery(name string, page int) (g Group) {
	var url strings.Builder
	url.WriteString("dauserprofile/init/gallery?username=")
	url.WriteString(name)
	url.WriteString("&page=")
	url.WriteString(strconv.Itoa(page))
	url.WriteString("&deviations_limit=50&with_subfolders=false")

	ujson(url.String(), &g)
	return
}
