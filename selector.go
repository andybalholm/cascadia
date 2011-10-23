package cascadia

import (
	"fmt"
	"html"
	"os"
)

// the Selector type, and functions for creating them

// A Selector is a function which tells whether a node matches or not.
type Selector func(*html.Node) bool

// Compile parses a selector and returns, if successful, a Selector object
// that can be used to match against html.Node objects.
func Compile(sel string) (Selector, os.Error) {
	p := &parser{s: sel}
	compiled, err := p.parseTypeSelector() // TODO: more complicated selectors
	if err != nil {
		return nil, err
	}

	if p.i < len(sel) {
		return nil, fmt.Errorf("parsing %q: %d bytes left over", sel, len(sel)-p.i)
	}

	return compiled, nil
}

// MatchAll returns a slice of the nodes that match the selector,
// from n and its children.
func (s Selector) MatchAll(n *html.Node) (result []*html.Node) {
	if s(n) {
		result = append(result, n)
	}

	for _, child := range n.Child {
		result = append(result, s.MatchAll(child)...)
	}

	return
}

// typeSelector returns a Selector that matches nodes with a given tag name.
func typeSelector(tag string) Selector {
	tag = toLowerASCII(tag)
	return func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == tag
	}
}

// toLowerASCII returns s with all ASCII capital letters lowercased.
func toLowerASCII(s string) string {
	var b []byte
	for i := 0; i < len(s); i++ {
		if c := s[i]; 'A' <= c && c <= 'Z' {
			if b == nil {
				b = make([]byte, len(s))
				copy(b, s)
			}
			b[i] = s[i] + ('a' - 'A')
		}
	}

	if b == nil {
		return s
	}

	return string(b)
}
