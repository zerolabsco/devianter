package devianter

import "testing"

// Regression: the shape check used to read m[0] and m[len(m)-1] without a length
// check, so a comment with an empty markup body panicked with index out of range
// and killed the caller's process.
func TestFlattenMarkupEmptyMarkup(t *testing.T) {
	if got := flattenMarkup(""); got != "" {
		t.Errorf("want an empty comment for empty markup, got %q", got)
	}
}

// tiptap is what DeviantArt actually serves today; the fixtures below are shaped
// like responses captured from the live API.
func TestFlattenMarkupTiptap(t *testing.T) {
	one := `{"version":1,"document":{"type":"doc","content":[` +
		`{"type":"paragraph","attrs":{"textAlign":"left"},"content":[` +
		`{"type":"text","text":"Nice artwork!"}]}]}}`
	if got := flattenMarkup(one); got != "Nice artwork!" {
		t.Errorf("want the paragraph's text, got %q", got)
	}

	// Regression: DeviantArt sends "version" as a number on some bodies and a
	// string on others. Modelling it as either type fails to unmarshal half of
	// them, so it must not be modelled at all.
	stringVersion := `{"version":"1","document":{"type":"doc","content":[` +
		`{"type":"paragraph","content":[{"type":"text","text":"hello"}]}]}}`
	if got := flattenMarkup(stringVersion); got != "hello" {
		t.Errorf(`want a string "version" handled the same as a numeric one, got %q`, got)
	}

	// Each block is its own line.
	two := `{"version":1,"document":{"type":"doc","content":[` +
		`{"type":"paragraph","content":[{"type":"text","text":"first"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"second"}]}]}}`
	if got, want := flattenMarkup(two), "first\nsecond"; got != want {
		t.Errorf("want one line per block:\n got %q\nwant %q", got, want)
	}

	// An empty paragraph is a blank line, not something to skip.
	blank := `{"version":1,"document":{"type":"doc","content":[` +
		`{"type":"paragraph","content":[{"type":"text","text":"a"}]},` +
		`{"type":"paragraph"},` +
		`{"type":"paragraph","content":[{"type":"text","text":"b"}]}]}}`
	if got, want := flattenMarkup(blank), "a\n\nb"; got != want {
		t.Errorf("want an empty paragraph preserved as a blank line:\n got %q\nwant %q", got, want)
	}

	// A hard break is a newline within its block.
	brk := `{"version":1,"document":{"type":"doc","content":[` +
		`{"type":"paragraph","content":[{"type":"text","text":"up"},` +
		`{"type":"hardBreak"},{"type":"text","text":"down"}]}]}}`
	if got, want := flattenMarkup(brk), "up\ndown"; got != want {
		t.Errorf("want a hardBreak as a newline:\n got %q\nwant %q", got, want)
	}

	// Marked-up runs are separate text nodes and must be concatenated, not
	// separated.
	marks := `{"version":1,"document":{"type":"doc","content":[` +
		`{"type":"paragraph","content":[` +
		`{"type":"text","text":"plain "},` +
		`{"type":"text","marks":[{"type":"bold"}],"text":"bold"},` +
		`{"type":"text","text":" tail"}]}]}}`
	if got, want := flattenMarkup(marks), "plain bold tail"; got != want {
		t.Errorf("want marked runs concatenated:\n got %q\nwant %q", got, want)
	}

	// Text carrying a link mark still surfaces; the href does not.
	link := `{"version":1,"document":{"type":"doc","content":[` +
		`{"type":"paragraph","content":[{"type":"text","marks":[` +
		`{"type":"link","attrs":{"href":"https://example.com"}}],"text":"click here"}]}]}}`
	if got, want := flattenMarkup(link), "click here"; got != want {
		t.Errorf("want the link text without the href:\n got %q\nwant %q", got, want)
	}

	// Headings are blocks like any other.
	heading := `{"version":1,"document":{"type":"doc","content":[` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"Title"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"body"}]}]}}`
	if got, want := flattenMarkup(heading), "Title\nbody"; got != want {
		t.Errorf("want a heading flattened as a block:\n got %q\nwant %q", got, want)
	}

	// Bodies carry HTML entities on the wire; plain text should not.
	entity := `{"version":1,"document":{"type":"doc","content":[` +
		`{"type":"paragraph","content":[{"type":"text","text":"I&#8217;d love &amp; more"}]}]}}`
	if got, want := flattenMarkup(entity), "I’d love & more"; got != want {
		t.Errorf("want HTML entities decoded:\n got %q\nwant %q", got, want)
	}

	// A node carrying no text of its own contributes nothing.
	emote := `{"version":1,"document":{"type":"doc","content":[` +
		`{"type":"paragraph","content":[{"type":"text","text":"hi "},` +
		`{"type":"da-emote","attrs":{"name":":happy:"}}]}]}}`
	if got, want := flattenMarkup(emote), "hi "; got != want {
		t.Errorf("want a textless node to contribute nothing:\n got %q\nwant %q", got, want)
	}
}

// Draft.js is legacy but still served on old bodies, so the path stays.
func TestFlattenMarkupDraftJS(t *testing.T) {
	draft := `{"blocks":[{"text":"hello there"}]}`
	if got := flattenMarkup(draft); got != "hello there" {
		t.Errorf("want the Draft.js block text, got %q", got)
	}

	// Regression: the block loop used to assign rather than accumulate, so every
	// block but the last was silently dropped and a multi-paragraph comment came
	// back as its closing line only.
	multi := `{"blocks":[{"text":"first"},{"text":"second"},{"text":"third"}]}`
	if got, want := flattenMarkup(multi), "first\nsecond\nthird"; got != want {
		t.Errorf("want every block, one per line:\n got %q\nwant %q", got, want)
	}

	// An empty block is a blank line, not something to skip.
	blank := `{"blocks":[{"text":"first"},{"text":""},{"text":"third"}]}`
	if got, want := flattenMarkup(blank), "first\n\nthird"; got != want {
		t.Errorf("want an empty block preserved as a blank line:\n got %q\nwant %q", got, want)
	}
}

func TestFlattenMarkupPassthrough(t *testing.T) {
	// An older, plain-HTML body passes through untouched.
	html := "<b>hello</b> there"
	if got := flattenMarkup(html); got != html {
		t.Errorf("want plain HTML passed through, got %q", got)
	}

	// Brace-shaped markup in no known format falls back to itself rather than to
	// an empty string.
	if got := flattenMarkup("{}"); got != "{}" {
		t.Errorf("want the original markup when nothing parses, got %q", got)
	}

	// Well-formed JSON that is neither format is still not silently eaten.
	other := `{"something":"else"}`
	if got := flattenMarkup(other); got != other {
		t.Errorf("want unrecognised JSON passed through, got %q", got)
	}

	// A single brace satisfies neither end of the shape check.
	if got := flattenMarkup("{"); got != "{" {
		t.Errorf("want a lone brace passed through, got %q", got)
	}
}
