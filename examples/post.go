package main

import (
	"git.macaw.me/skunky/devianter"
)

func main() {
	id := "973578309"
	d := devianter.Deviation(id, "Thrumyeye")

	println("Post Name:", d.Deviation.Title, "\nIMG url:", d.IMG)

	c := devianter.Comments(id, "", 0, 1)
	println("\n\nPost Comments:", c.Total)

	for _, a := range c.Thread {
		if a.User.IsBanned {
			a.User.Username += " [v bane]"
		}
		println(a.User.Username+":", a.Comment)
	}

	search := devianter.Search("skunk", "2", 'a')
	println(search.Total, search.Pages)
}
