package cascadia

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestLangInheritance(t *testing.T) {
	for _, tt := range []struct {
		name, source, selector string
		want                   int
	}{
		{"inherited", `<div lang="en"><p>text</p></div>`, "p:lang(en)", 1},
		{"different language", `<div lang="en"><p lang="fr">text</p></div>`, "p:lang(en)", 0},
		{"ancestor override", `<div lang="en"><section lang="fr"><p>text</p></section></div>`, "p:lang(en)", 0},
		{"empty language", `<div lang="en"><p lang="">text</p></div>`, "p:lang(en)", 0},
		{"empty ancestor", `<div lang="en"><section lang=""><p>text</p></section></div>`, "p:lang(en)", 0},
		{"nearest match", `<div lang="en"><section lang="fr"><p>text</p></section></div>`, "p:lang(fr)", 1},
		{"subtag", `<div lang="en-GB"><p>text</p></div>`, "p:lang(en)", 1},
		{"no language", `<p>text</p>`, "p:lang(en)", 0},
		{"elements only", `<p lang="en">text<!--comment--></p>`, ":lang(en)", 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := html.Parse(strings.NewReader(tt.source))
			if err != nil {
				t.Fatal(err)
			}
			sel, err := Parse(tt.selector)
			if err != nil {
				t.Fatal(err)
			}
			if got := len(QueryAll(doc, sel)); got != tt.want {
				t.Errorf("QueryAll(%s, %s) matched %d nodes, want %d", tt.source, tt.selector, got, tt.want)
			}
		})
	}
}
