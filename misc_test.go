package devianter

import "testing"

// Regression: AEmedia used to call log.Fatalln on an unknown type, terminating
// the calling process. If this test ever regresses it will not fail — the test
// binary will exit(1) mid-run.
func TestAEmediaInvalidTypeReturnsError(t *testing.T) {
	_, err := AEmedia("someuser", 'z')
	if err == nil {
		t.Fatal("want an error for an unknown type, got nil")
	}
}

// The argument checks must run in an order that never leaves a valid-looking
// call unreported: a short name is an error regardless of type.
func TestAEmediaShortNameReturnsError(t *testing.T) {
	if _, err := AEmedia("", 'a'); err == nil {
		t.Fatal("want an error for an empty name, got nil")
	}
}

// Regression: PerformSearch used to call log.Fatalln on an unknown scope. As
// above, a regression exits the test binary rather than failing this test.
func TestPerformSearchInvalidScopeReturnsError(t *testing.T) {
	_, _, err := PerformSearch("cats", 0, 'z')
	if err == nil {
		t.Fatal("want an error for an unknown scope, got nil")
	}
}

// The account-scoped searches need a username, and must say so rather than
// requesting a URL with an empty one.
func TestPerformSearchAccountScopesRequireUser(t *testing.T) {
	for _, scope := range []rune{'g', 'f'} {
		if _, _, err := PerformSearch("cats", 0, scope); err == nil {
			t.Errorf("want an error for scope %q with no username, got nil", scope)
		}
	}
}
