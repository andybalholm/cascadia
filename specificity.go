package cascadia

import (
	"golang.org/x/net/html"
)

// Specificity is the CSS specificity as defined in
// https://www.w3.org/TR/selectors/#specificity-rules
// with the convention Specificity = [A,B,C].
type Specificity [3]uint8

// returns `true` if s < other (strictly), false otherwise
func (s Specificity) less(other Specificity) bool {
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
		if out.less(sel.Specificity) {
			out = sel.Specificity
		}
	}
	return out
}

// MatchWithSpecificity return `true` if `element` matches `s`.
// In this case, the greatest specificity (among the sub-selectors matching `element`)
// is returned.
// From https://www.w3.org/TR/selectors/#specificity-rules
// " If the selector is a selector list, this number is calculated for each selector in the list.
// For a given matching process against the list, the specificity in effect is that of the most
// specific selector in the list that matches. "
func (s Selector) MatchWithSpecificity(element *html.Node) (bool, Specificity) {
	var (
		maxSpec Specificity
		found   bool
	)
	for _, sel := range s {
		if sel.match(element) {
			found = true
			if maxSpec.less(sel.Specificity) {
				maxSpec = sel.Specificity
			}
		}
	}
	return found, maxSpec
}
