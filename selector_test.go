package cascadia

import (
	"html"
	"strings"
	"testing"
)

type selectorTest struct {
	HTML, selector string
	results        []string
}

func nodeString(n *html.Node) string {
	switch n.Type {
	case html.TextNode:
		return n.Data
	case html.ElementNode:
		return html.Token{
			Type: html.StartTagToken,
			Data: n.Data,
			Attr: n.Attr,
		}.String()
	}
	return ""
}

var selectorTests = []selectorTest{
	{
		`<body><address>This address...</address></body>`,
		"address",
		[]string{
			"<address>",
		},
	},
	{
		`<html><head></head><body></body></html>`,
		"*",
		[]string{
			"",
			"<html>",
			"<head>",
			"<body>",
		},
	},
	{
		`<p id="foo"><p id="bar">`,
		"#foo",
		[]string{
			`<p id="foo">`,
		},
	},
	{
		`<ul><li id="t1"><p id="t1">`,
		"li#t1",
		[]string{
			`<li id="t1">`,
		},
	},
	{
		`<ol><li id="t4"><li id="t44">`,
		"*#t4",
		[]string{
			`<li id="t4">`,
		},
	},
}

func TestSelectors(t *testing.T) {
	for _, test := range selectorTests {
		s, err := Compile(test.selector)
		if err != nil {
			t.Errorf("error compiling %q: %s", test.selector, err)
			continue
		}

		doc, err := html.Parse(strings.NewReader(test.HTML))
		if err != nil {
			t.Errorf("error parsing %q: %s", test.HTML, err)
			continue
		}

		matches := s.MatchAll(doc)
		if len(matches) != len(test.results) {
			t.Errorf("wanted %d elements, got %d instead", len(test.results), len(matches))
			continue
		}

		for i, m := range matches {
			got := nodeString(m)
			if got != test.results[i] {
				t.Errorf("wanted %s, got %s instead", test.results[i], got)
			}
		}
	}
}
