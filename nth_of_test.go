package cascadia

import (
	"reflect"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

const nthOfTestHTML = `<html><body>
<ul><li id="l1" class="a">1</li><li id="l2">2</li><li id="l3" class="a">3</li><li id="l4" class="a b">4</li><li id="l5">5</li></ul>
<div id="d"><p id="p1" class="a">x</p><span id="s1">y</span><p id="p2">z</p></div>
</body></html>`

func TestNthOfMatch(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(nthOfTestHTML))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct{ sel, ids string }{
		{":nth-child(2 of .a)", "l3"},
		{"li:nth-child(odd of .a)", "l1 l4"},
		{":nth-child(n of .a)", "l1 l3 l4 p1"},
		{":nth-child(-n+2 of .a)", "l1 l3 p1"},
		{":nth-last-child(1 of .a)", "l4 p1"},
		{":nth-last-child(2 of .a)", "l3"},
		{":nth-child(1 of .b)", "l4"},
		{"li:nth-child(1 of .b, #l5)", "l4"},
		// the element itself has to match S
		{".a:nth-child(2 of li)", ""},
		{":nth-child(2 of li):not(.a)", "l2"},
		{":nth-child(2 of p, span)", "s1"},
		{":nth-child( 2  OF  .a )", "l3"},
		{":nth-child(2 of /* c */ .a)", "l3"},
		{"ul > :nth-child(1 of :not(.a))", "l2"},
	}
	for _, tc := range tests {
		if got := matchIDs(t, doc, tc.sel); got != tc.ids {
			t.Errorf("%s: got %q, want %q", tc.sel, got, tc.ids)
		}
	}
}

func TestNthOfSerialize(t *testing.T) {
	tests := map[string]string{
		":nth-child(2 of .a)":              ":nth-child(0n+2 of .a)",
		":nth-last-child(odd of li.a, #x)": ":nth-last-child(2n+1 of li.a, #x)",
		":nth-child(1 of p)":               ":nth-child(0n+1 of p)",
		"li:nth-child(-n+3 of .a > span)":  "li:nth-child(-1n+3 of .a > span)",
	}
	for sel, want := range tests {
		s, err := ParseGroup(sel)
		if err != nil {
			t.Fatalf("error compiling %q: %s", sel, err)
		}
		got := s.String()
		if got != want {
			t.Errorf("%s: serialized as %q, want %q", sel, got, want)
		}
		s2, err := ParseGroup(got)
		if err != nil {
			t.Fatalf("error compiling serialized %q: %s", got, err)
		}
		if !reflect.DeepEqual(s, s2) {
			t.Errorf("%s: round trip through %q changed the selector", sel, got)
		}
	}
}

func TestNthOfSpecificity(t *testing.T) {
	tests := map[string]Specificity{
		":nth-child(2 of .a)":         {0, 2, 0},
		":nth-child(2 of .a, #x)":     {1, 1, 0},
		"li:nth-child(2 of .a)":       {0, 2, 1},
		":nth-child(2 of li)":         {0, 1, 1},
		":nth-last-child(2 of p, .a)": {0, 2, 0},
	}
	for sel, want := range tests {
		s, err := Parse(sel)
		if err != nil {
			t.Fatalf("error compiling %q: %s", sel, err)
		}
		if got := s.Specificity(); got != want {
			t.Errorf("%s: specificity %v, want %v", sel, got, want)
		}
	}
}

func TestNthOfInvalid(t *testing.T) {
	for _, sel := range []string{
		":nth-child(2 of)",
		":nth-child(2 of )",
		":nth-child(2 off p)",
		":nth-child(2 of p q, )",
		":nth-of-type(2 of p)",
		":nth-last-of-type(1 of p)",
		":nth-child(of p)",
		":nth-child(2of p)",
		":nth-child(2 ofp)",
		":nth-child(2 of ::before)",
	} {
		if _, err := ParseGroup(sel); err == nil {
			t.Errorf("%s: expected an error", sel)
		}
	}
}
