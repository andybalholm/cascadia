package cascadia

import "golang.org/x/net/html"

// This file implements Matcher interface for basic selectors.

type matcher = func(*html.Node) bool

type tagSelector struct {
	tag string
}

type classSelector struct {
	class string
}

type idSelector struct {
	id string
}

type attrSelector struct {
	match matcher
}

type pseudoClassSelector struct {
	match matcher

	// The specificty of a pseudo-class is not constant.
	// See https://www.w3.org/TR/selectors/#specificity-rules for the special cases
	specificity Specificity
}

type compoundSelector struct {
	selectors     []Matcher
	pseudoElement string
}

func (c compoundSelector) PseudoElement() string {
	return c.pseudoElement
}

type combinedSelector struct {
	first      Matcher
	combinator byte
	second     Matcher

	pseudoElement string
}

func (c combinedSelector) PseudoElement() string {
	return c.pseudoElement
}

// Matches elements with a given tag name.
func (t tagSelector) Match(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == t.tag
}

// Matches elements by id attribute.
func (t idSelector) Match(n *html.Node) bool {
	return matchAttribute(n, "id", func(s string) bool {
		return s == t.id
	})
}

// Matches elements by id attribute.
func (t classSelector) Match(n *html.Node) bool {
	return matchAttribute(n, "class", func(s string) bool {
		return matchInclude(t.class, s)
	})
}

// Matches elements by attribute value.
func (t attrSelector) Match(n *html.Node) bool {
	return t.match(n)
}

// Matches elements by attribute value.
func (t pseudoClassSelector) Match(n *html.Node) bool {
	return t.match(n)
}

// Matches elements if each sub-selectors matches.
func (t compoundSelector) Match(n *html.Node) bool {
	if len(t.selectors) == 0 {
		return n.Type == html.ElementNode
	}

	for _, sel := range t.selectors {
		if !sel.Match(n) {
			return false
		}
	}
	return true
}

func (t combinedSelector) Match(n *html.Node) bool {
	if t.first == nil {
		return false // maybe we should panic
	}
	switch t.combinator {
	case 0:
		return t.first.Match(n)
	case ' ':
		return descendantMatch(t.first, t.second, n)
	case '>':
		return childMatch(t.first, t.second, n)
	case '+':
		return siblingMatch(t.first, t.second, true, n)
	case '~':
		return siblingMatch(t.first, t.second, false, n)
	default:
		panic("unknown combinator")
	}
}
