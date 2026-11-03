package cascadia

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

const isTestHTML = `<html><body>
<div id="d1" class="a"><p id="p1" class="x">1</p><span id="s1" class="y">2</span></div>
<div id="d2" class="b"><p id="p2">3</p><em id="e1" class="x">4</em></div>
<section id="sec"><p id="p3" class="y">5</p></section>
</body></html>`

func matchIDs(t *testing.T, doc *html.Node, sel string) string {
	t.Helper()
	s, err := ParseGroupWithPseudoElements(sel)
	if err != nil {
		t.Fatalf("error compiling %q: %s", sel, err)
	}
	var ids []string
	for _, n := range QueryAll(doc, s) {
		for _, a := range n.Attr {
			if a.Key == "id" {
				ids = append(ids, a.Val)
			}
		}
	}
	return strings.Join(ids, " ")
}

func TestIsWhereMatch(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(isTestHTML))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct{ sel, ids string }{
		{":is(.x, .y)", "p1 s1 e1 p3"},
		{":where(.x, .y)", "p1 s1 e1 p3"},
		{"p:is(.x, .y)", "p1 p3"},
		{":is(p, em).x", "p1 e1"},
		{":is(.a, .b) > .x", "p1 e1"},
		{":is(div, section) p", "p1 p2 p3"},
		{":is(div.a p)", "p1"},
		{"div :is(p, span):not(.y)", "p1 p2"},
		{":is(:is(#p1, #p2), :where(#p3))", "p1 p2 p3"},
		{":is(#d1) :is(.x)", "p1"},
		// forgiving: invalid arguments are dropped
		{":is(.x, :unknown-pseudo, p::before, [=x], #p3)", "p1 e1 p3"},
		{":is(:unknown(a, (b)), #s1)", "s1"},
		{":is()", ""},
		{":where(:unknown)", ""},
		{":is( .x , .y )", "p1 s1 e1 p3"},
	}
	for _, tc := range tests {
		if got := matchIDs(t, doc, tc.sel); got != tc.ids {
			t.Errorf("%s: got %q, want %q", tc.sel, got, tc.ids)
		}
	}
}

func TestIsWhereSpecificity(t *testing.T) {
	tests := []struct {
		sel    string
		spec   Specificity
		pseudo string
	}{
		{":is(p, .x, #id)", Specificity{1, 0, 0}, ""},
		{":is(p, div)", Specificity{0, 0, 1}, ""},
		{"p:is(.a.b, .c)", Specificity{0, 2, 1}, ""},
		{":where(#id, .x)", Specificity{0, 0, 0}, ""},
		{"p:where(#id).x", Specificity{0, 1, 1}, ""},
		{":is(:where(#id), .x)", Specificity{0, 1, 0}, ""},
		{":is(.fas, .far, .fa)::before", Specificity{0, 1, 1}, "before"},
		{":where(.fa)::before", Specificity{0, 0, 1}, "before"},
		{":is()", Specificity{0, 0, 0}, ""},
	}
	for _, tc := range tests {
		s, err := ParseWithPseudoElement(tc.sel)
		if err != nil {
			t.Errorf("error compiling %q: %s", tc.sel, err)
			continue
		}
		if s.Specificity() != tc.spec {
			t.Errorf("%s: specificity %v, want %v", tc.sel, s.Specificity(), tc.spec)
		}
		if s.PseudoElement() != tc.pseudo {
			t.Errorf("%s: pseudo-element %q, want %q", tc.sel, s.PseudoElement(), tc.pseudo)
		}
	}
}

func TestIsWhereSerialize(t *testing.T) {
	tests := []struct{ sel, want string }{
		{":is(.x, p)", ":is(.x, p)"},
		{":where(div > p)", ":where(div > p)"},
		{":is(.x, :unknown, p)", ":is(.x, p)"},
		{":is()", ":is()"},
	}
	for _, tc := range tests {
		s, err := Parse(tc.sel)
		if err != nil {
			t.Errorf("error compiling %q: %s", tc.sel, err)
			continue
		}
		if got := s.String(); got != tc.want {
			t.Errorf("%s: serialized %q, want %q", tc.sel, got, tc.want)
		}
	}
}

func TestIsWhereInvalid(t *testing.T) {
	for _, sel := range []string{":is", ":is(.x", ":where(.x, (", ":is(\"a)", ":is(.x, [, #p3)"} {
		if _, err := Parse(sel); err == nil {
			t.Errorf("%q should not compile", sel)
		}
	}
}
