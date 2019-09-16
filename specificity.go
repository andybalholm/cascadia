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
		if out.Less(sel.Specificity) {
			out = sel.Specificity
		}
	}
	return out
}

type MatchDetail struct {
	PseudoElement string
	Specificity   Specificity
}

// MatchDetails return `true` if `element` matches `s`.
// In this case, the list of specificity and optionnal pseudoElement matching is returned.
func (s Selector) MatchDetails(element *html.Node) (bool, []MatchDetail) {
	var out []MatchDetail
	for _, sel := range s {
		if sel.match(element) {
			out = append(out, MatchDetail{PseudoElement: sel.PseudoElement, Specificity: sel.Specificity})
		}
	}
	return len(out) > 0, out
}
