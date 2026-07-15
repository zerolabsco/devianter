package devianter

import "testing"

// Regression: flattenComment's shape check used to read m[0] and m[len(m)-1]
// without a length check, so a comment with an empty markup body panicked with
// index out of range and killed the caller's process.
func TestFlattenCommentEmptyMarkup(t *testing.T) {
	if got := flattenComment(""); got != "" {
		t.Errorf("want an empty comment for empty markup, got %q", got)
	}
}

func TestFlattenComment(t *testing.T) {
	// A newer, Draft.js-encoded body is flattened to its text.
	draft := `{"blocks":[{"text":"hello there"}]}`
	if got := flattenComment(draft); got != "hello there" {
		t.Errorf("want the Draft.js block text, got %q", got)
	}

	// An older, plain-HTML body passes through untouched.
	html := "<b>hello</b> there"
	if got := flattenComment(html); got != html {
		t.Errorf("want plain HTML passed through, got %q", got)
	}

	// Regression: the block loop used to assign rather than accumulate, so every
	// block but the last was silently dropped and a multi-paragraph comment came
	// back as its closing line only.
	multi := `{"blocks":[{"text":"first"},{"text":"second"},{"text":"third"}]}`
	if got, want := flattenComment(multi), "first\nsecond\nthird"; got != want {
		t.Errorf("want every block, one per line:\n got %q\nwant %q", got, want)
	}

	// An empty block is a blank line in the comment, not something to skip.
	blank := `{"blocks":[{"text":"first"},{"text":""},{"text":"third"}]}`
	if got, want := flattenComment(blank), "first\n\nthird"; got != want {
		t.Errorf("want an empty block preserved as a blank line:\n got %q\nwant %q", got, want)
	}

	// Brace-shaped markup that isn't a Draft.js document falls back to itself
	// rather than to an empty string.
	if got := flattenComment("{}"); got != "{}" {
		t.Errorf("want the original markup when there are no blocks, got %q", got)
	}

	// A single brace satisfies neither end of the shape check.
	if got := flattenComment("{"); got != "{" {
		t.Errorf("want a lone brace passed through, got %q", got)
	}
}
