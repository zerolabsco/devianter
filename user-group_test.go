package devianter

import (
	"testing"
	"time"
)

// Regression: Favourites and Gallery took folderid as a variadic, which invites
// omitting it, but then indexed folderid[0] with no length check — so the call
// the signature invites most panicked with index out of range.
//
// Both reach a request before returning, and neither takes a base URL, so these
// squeeze Timeout down to make that request fail immediately. The failure is
// expected and ignored: only the absence of a panic is under test.
func TestFolderidIsOptional(t *testing.T) {
	defer func(d time.Duration) { Timeout = d }(Timeout)
	Timeout = time.Millisecond

	s := Group{Name: "someuser"}

	t.Run("Gallery", func(t *testing.T) {
		if _, _, err := s.Gallery(0); err != nil {
			t.Errorf("omitting folderid is not an argument error, got %v", err)
		}
	})

	t.Run("Favourites all", func(t *testing.T) {
		_, _ = s.Favourites(0, true)
	})

	t.Run("Favourites default listing", func(t *testing.T) {
		_, _ = s.Favourites(0, false)
	})
}

// Name is what every request is scoped to, so Gallery reports its absence
// rather than requesting a URL with an empty username.
func TestGalleryRequiresName(t *testing.T) {
	var s Group
	if _, _, err := s.Gallery(0, 0); err == nil {
		t.Fatal("want an error for a Group with no Name, got nil")
	}
}
