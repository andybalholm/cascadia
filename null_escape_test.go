package cascadia

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestNullEscapeUsesReplacementCharacter(t *testing.T) {
	document, err := html.Parse(strings.NewReader(`<p id="&#0;" class="&#0;">text</p>`))
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{`#\0`, `#\000000`, `#\0 `, `.\0`, `[id="\0"]`} {
		t.Run(input, func(t *testing.T) {
			selector, err := Parse(input)
			if err != nil {
				t.Fatal(err)
			}
			matches := QueryAll(document, selector)
			if len(matches) != 1 || matches[0].Data != "p" {
				t.Fatalf("got %d matches, want the paragraph with replacement character", len(matches))
			}
		})
	}
}

func TestEscapedCodePointControls(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{`\41`, "A"}, {`\10ffff`, "\U0010ffff"}, {`\110000`, "\ufffd"}, {`\d800`, "\ufffd"}, {`\000041B`, "A"},
	} {
		p := parser{s: tc.input}
		got, err := p.parseEscape()
		if err != nil || got != tc.want {
			t.Errorf("%q: got %q, %v; want %q", tc.input, got, err, tc.want)
		}
	}
}
