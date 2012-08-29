package cascadia

import (
	"exp/html"
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
	{
		`<ul><li class="t1"><li class="t2">`,
		".t1",
		[]string{
			`<li class="t1">`,
		},
	},
	{
		`<p class="t1 t2">`,
		"p.t1",
		[]string{
			`<p class="t1 t2">`,
		},
	},
	{
		`<div class="test">`,
		"div.teST",
		[]string{},
	},
	{
		`<p class="t1 t2">`,
		".t1.fail",
		[]string{},
	},
	{
		`<p class="t1 t2">`,
		"p.t1.t2",
		[]string{
			`<p class="t1 t2">`,
		},
	},
	{
		`<p><p title="title">`,
		"p[title]",
		[]string{
			`<p title="title">`,
		},
	},
	{
		`<address><address title="foo"><address title="bar">`,
		`address[title="foo"]`,
		[]string{
			`<address title="foo">`,
		},
	},
	{
		`<p title="tot foo bar">`,
		`[    	title        ~=       foo    ]`,
		[]string{
			`<p title="tot foo bar">`,
		},
	},
	{
		`<p title="hello world">`,
		`[title~="hello world"]`,
		[]string{},
	},
	{
		`<p lang="en"><p lang="en-gb"><p lang="enough"><p lang="fr-en">`,
		`[lang|="en"]`,
		[]string{
			`<p lang="en">`,
			`<p lang="en-gb">`,
		},
	},
	{
		`<p title="foobar"><p title="barfoo">`,
		`[title^="foo"]`,
		[]string{
			`<p title="foobar">`,
		},
	},
	{
		`<p title="foobar"><p title="barfoo">`,
		`[title$="bar"]`,
		[]string{
			`<p title="foobar">`,
		},
	},
	{
		`<p title="foobarufoo">`,
		`[title*="bar"]`,
		[]string{
			`<p title="foobarufoo">`,
		},
	},
	{
		`<p class="t1 t2">`,
		".t1:not(.t2)",
		[]string{},
	},
	{
		`<div class="t3">`,
		`div:not(.t1)`,
		[]string{
			`<div class="t3">`,
		},
	},
	{
		`<ol><li id=1><li id=2><li id=3></ol>`,
		`li:nth-child(odd)`,
		[]string{
			`<li id="1">`,
			`<li id="3">`,
		},
	},
	{
		`<ol><li id=1><li id=2><li id=3></ol>`,
		`li:nth-child(even)`,
		[]string{
			`<li id="2">`,
		},
	},
	{
		`<ol><li id=1><li id=2><li id=3></ol>`,
		`li:nth-child(-n+2)`,
		[]string{
			`<li id="1">`,
			`<li id="2">`,
		},
	},
	{
		`<ol><li id=1><li id=2><li id=3></ol>`,
		`li:nth-child(3n+1)`,
		[]string{
			`<li id="1">`,
		},
	},
	{
		`<ol><li id=1><li id=2><li id=3><li id=4></ol>`,
		`li:nth-last-child(odd)`,
		[]string{
			`<li id="2">`,
			`<li id="4">`,
		},
	},
	{
		`<ol><li id=1><li id=2><li id=3><li id=4></ol>`,
		`li:nth-last-child(even)`,
		[]string{
			`<li id="1">`,
			`<li id="3">`,
		},
	},
	{
		`<ol><li id=1><li id=2><li id=3><li id=4></ol>`,
		`li:nth-last-child(-n+2)`,
		[]string{
			`<li id="3">`,
			`<li id="4">`,
		},
	},
	{
		`<ol><li id=1><li id=2><li id=3><li id=4></ol>`,
		`li:nth-last-child(3n+1)`,
		[]string{
			`<li id="1">`,
			`<li id="4">`,
		},
	},
	{
		`<p>some text <span id="1">and a span</span><span id="2"> and another</span></p>`,
		`span:first-child`,
		[]string{
			`<span id="1">`,
		},
	},
	{
		`<span>a span</span> and some text`,
		`span:last-child`,
		[]string{
			`<span>`,
		},
	},
	{
		`<address></address><p id=1><p id=2>`,
		`p:nth-of-type(2)`,
		[]string{
			`<p id="2">`,
		},
	},
	{
		`<address></address><p id=1><p id=2></p><a>`,
		`p:nth-last-of-type(2)`,
		[]string{
			`<p id="1">`,
		},
	},
	{
		`<address></address><p id=1><p id=2></p><a>`,
		`p:last-of-type`,
		[]string{
			`<p id="2">`,
		},
	},
	{
		`<address></address><p id=1><p id=2></p><a>`,
		`p:first-of-type`,
		[]string{
			`<p id="1">`,
		},
	},
	{
		`<div><p id="1"></p><a></a></div><div><p id="2"></p></div>`,
		`p:only-child`,
		[]string{
			`<p id="2">`,
		},
	},
	{
		`<div><p id="1"></p><a></a></div><div><p id="2"></p><p id="3"></p></div>`,
		`p:only-of-type`,
		[]string{
			`<p id="1">`,
		},
	},
	{
		`<p id="1"><!-- --><p id="2">Hello<p id="3"><span>`,
		`:empty`,
		[]string{
			`<head>`,
			`<p id="1">`,
			`<span>`,
		},
	},
	{
		`<div><p id="1"><table><tr><td><p id="2"></table></div><p id="3">`,
		`div p`,
		[]string{
			`<p id="1">`,
			`<p id="2">`,
		},
	},
	{
		`<div><p id="1"><table><tr><td><p id="2"></table></div><p id="3">`,
		`div table p`,
		[]string{
			`<p id="2">`,
		},
	},
	{
		`<div><p id="1"><div><p id="2"></div><table><tr><td><p id="3"></table></div>`,
		`div > p`,
		[]string{
			`<p id="1">`,
			`<p id="2">`,
		},
	},
	{
		`<p id="1"><p id="2"></p><address></address><p id="3">`,
		`p ~ p`,
		[]string{
			`<p id="2">`,
			`<p id="3">`,
		},
	},
	{
		`<p id="1"><p id="2"></p><address></address><p id="3">`,
		`p + p`,
		[]string{
			`<p id="2">`,
		},
	},
	{
		`<ul><li></li><li></li></ul><p>`,
		`li, p`,
		[]string{
			"<li>",
			"<li>",
			"<p>",
		},
	},
	{
		`<p id="1"><p id="2"></p><address></address><p id="3">`,
		`p +/*This is a comment*/ p`,
		[]string{
			`<p id="2">`,
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
