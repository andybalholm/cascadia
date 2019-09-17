package cascadia

import (
	"golang.org/x/net/html"
)

// Specificity is the CSS specificity as defined in
// https://www.w3.org/TR/selectors/#specificity-rules
// with the convention Specificity = [A,B,C].
type Specificity [3]uint8

// returns `true` if s < other (strictly), false otherwise
func (s Specificity) Less(other Specificity) bool {
	for i := range s {
		if s[i] < other[i] {
			return true
		}
		if s[i] > other[i] {
			return false
		}
	}
	return false
}

func (s Specificity) add(other Specificity) Specificity {
	for i, sp := range other {
		s[i] += sp
	}
	return s
}

func (s Selector) maximumSpecificity() Specificity {
	var out Specificity
	for _, sel := range s {
		sp := sel.Specificity()
		if out.Less(sp) {
			out = sp
		}
	}
	return out
}

// basic specificity
func (c tagSelector) Specificity() Specificity {
	return Specificity{0, 0, 1}
}

func (c classSelector) Specificity() Specificity {
	return Specificity{0, 1, 0}
}

func (c idSelector) Specificity() Specificity {
	return Specificity{1, 0, 0}
}

func (c attrSelector) Specificity() Specificity {
	return Specificity{0, 1, 0}
}

func (c pseudoClassSelector) Specificity() Specificity {
	return c.specificity
}

func (c compoundSelector) Specificity() Specificity {
	var out Specificity
	for _, sel := range c.selectors {
		out = out.add(sel.Specificity())
	}
	if c.pseudoElement != "" {
		out = out.add(Specificity{0, 0, 1})
	}
	return out
}

func (c combinedSelector) Specificity() Specificity {
	s := c.first.Specificity()
	if c.second != nil {
		s = s.add(c.second.Specificity())
	}
	return s
}

type MatchDetail struct {
	PseudoElement string
	Specificity   Specificity
}

// MatchDetails return the list of specificity and optionnal pseudoElement of
// matching selectors.
func (s Selector) MatchDetails(element *html.Node) []MatchDetail {
	var out []MatchDetail
	for _, sel := range s {
		if sel.Match(element) {
			out = append(out, MatchDetail{PseudoElement: sel.PseudoElement(), Specificity: sel.Specificity()})
		}
	}
	return out
}
