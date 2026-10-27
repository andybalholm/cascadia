package cascadia

import "golang.org/x/net/html"

// queryEvaluator evaluates the built-in selector tree while retaining caches
// whose lifetime is exactly one query. It deliberately falls back to Sel.Match
// for selector implementations that do not benefit from query-local state.
type queryEvaluator struct {
	document *IndexedDocument
	relative map[*relativeSelector]map[*html.Node]struct{}
}

type elementName struct {
	namespace string
	data      string
}

func newQueryEvaluator(document *IndexedDocument) *queryEvaluator {
	return &queryEvaluator{
		document: document,
		relative: make(map[*relativeSelector]map[*html.Node]struct{}),
	}
}

func (e *queryEvaluator) match(m Matcher, n *html.Node) bool {
	switch s := m.(type) {
	case SelectorGroup:
		return e.matchGroup(s, n)
	case Sel:
		return e.matchSel(s, n)
	default:
		return m.Match(n)
	}
}

func (e *queryEvaluator) matchGroup(group SelectorGroup, n *html.Node) bool {
	for _, sel := range group {
		if e.matchSel(sel, n) {
			return true
		}
	}
	return false
}

func (e *queryEvaluator) matchSel(sel Sel, n *html.Node) bool {
	switch s := sel.(type) {
	case compoundSelector:
		if len(s.selectors) == 0 {
			return n.Type == html.ElementNode
		}
		for _, part := range s.selectors {
			if !e.matchSel(part, n) {
				return false
			}
		}
		return true
	case combinedSelector:
		return e.matchCombined(s, n)
	case notPseudoClassSelector:
		return n.Type == html.ElementNode && !e.matchGroup(s.match, n)
	case relativePseudoClassSelector:
		return e.matchRelative(s, n)
	case nthPseudoClassSelector:
		return e.matchNth(s, n)
	case onlyChildPseudoClassSelector:
		return e.matchOnlyChild(s, n)
	default:
		return sel.Match(n)
	}
}

func (e *queryEvaluator) matchCombined(s combinedSelector, n *html.Node) bool {
	if s.first == nil {
		return false
	}
	if s.combinator == 0 {
		return e.matchSel(s.first, n)
	}
	if s.second == nil || !e.matchSel(s.second, n) {
		return false
	}
	switch s.combinator {
	case ' ':
		for parent := n.Parent; parent != nil; parent = parent.Parent {
			if e.matchSel(s.first, parent) {
				return true
			}
		}
	case '>':
		return n.Parent != nil && e.matchSel(s.first, n.Parent)
	case '+':
		for previous := n.PrevSibling; previous != nil; previous = previous.PrevSibling {
			if previous.Type == html.TextNode || previous.Type == html.CommentNode {
				continue
			}
			return e.matchSel(s.first, previous)
		}
	case '~':
		for previous := n.PrevSibling; previous != nil; previous = previous.PrevSibling {
			if e.matchSel(s.first, previous) {
				return true
			}
		}
	}
	return false
}

func (e *queryEvaluator) matchRelative(s relativePseudoClassSelector, n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	for i := range s.rels {
		relative := &s.rels[i]
		combinator := relative.combinator
		if combinator == 0 {
			if s.name == "haschild" {
				combinator = '>'
			} else {
				combinator = ' '
			}
		}
		if matched, handled := e.matchRelativeFromIndex(n, combinator, relative); handled {
			if matched {
				return true
			}
			continue
		}
		if e.matchRelativeSelector(n, combinator, relative.sel) {
			return true
		}
	}
	return false
}

func (e *queryEvaluator) matchRelativeFromIndex(n *html.Node, combinator byte, relative *relativeSelector) (bool, bool) {
	if e.document == nil || (combinator != ' ' && combinator != '>') {
		return false, false
	}
	matched, exists := e.relative[relative]
	if !exists {
		candidates, ok := e.document.bestCandidates(relative.sel)
		if !ok {
			return false, false
		}
		matched = make(map[*html.Node]struct{})
		for _, candidate := range candidates {
			if !e.matchSel(relative.sel, candidate) {
				continue
			}
			if combinator == '>' {
				if candidate.Parent != nil {
					matched[candidate.Parent] = struct{}{}
				}
				continue
			}
			for parent := candidate.Parent; parent != nil; parent = parent.Parent {
				matched[parent] = struct{}{}
			}
		}
		e.relative[relative] = matched
	}
	_, ok := matched[n]
	return ok, true
}

func (e *queryEvaluator) matchRelativeSelector(n *html.Node, combinator byte, sel Sel) bool {
	switch combinator {
	case ' ':
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			if e.matchSel(sel, child) || e.matchRelativeSelector(child, ' ', sel) {
				return true
			}
		}
	case '>':
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			if e.matchSel(sel, child) {
				return true
			}
		}
	case '+':
		for next := n.NextSibling; next != nil; next = next.NextSibling {
			if next.Type == html.ElementNode {
				return e.matchSel(sel, next)
			}
		}
	case '~':
		for next := n.NextSibling; next != nil; next = next.NextSibling {
			if next.Type == html.ElementNode && e.matchSel(sel, next) {
				return true
			}
		}
	}
	return false
}

func (e *queryEvaluator) matchNth(s nthPseudoClassSelector, n *html.Node) bool {
	if n.Type != html.ElementNode || n.Parent == nil {
		return false
	}
	metadata, ok := e.siblingMetadata(n)
	if !ok {
		return s.Match(n)
	}
	position := int(metadata.position)
	if s.ofType {
		position = int(metadata.typePosition)
	}
	if s.last {
		if s.ofType {
			position = int(metadata.typeReversePosition)
		} else {
			position = int(metadata.reversePosition)
		}
	}
	value := position - s.b
	if s.a == 0 {
		return value == 0
	}
	return value%s.a == 0 && value/s.a >= 0
}

func (e *queryEvaluator) matchOnlyChild(s onlyChildPseudoClassSelector, n *html.Node) bool {
	if n.Type != html.ElementNode || n.Parent == nil {
		return false
	}
	metadata, ok := e.siblingMetadata(n)
	if !ok {
		return s.Match(n)
	}
	if !s.ofType {
		return metadata.count == 1
	}
	return metadata.typeCount == 1
}

func (e *queryEvaluator) siblingMetadata(n *html.Node) (siblingMetadata, bool) {
	if e.document == nil {
		return siblingMetadata{}, false
	}
	position, ok := e.document.order[n]
	if !ok || position >= len(e.document.sibling) {
		return siblingMetadata{}, false
	}
	return e.document.sibling[position], true
}

func queryAllWithEvaluator(root *html.Node, matcher Matcher, evaluator *queryEvaluator) []*html.Node {
	var result []*html.Node
	var visit func(*html.Node)
	visit = func(parent *html.Node) {
		for child := parent.FirstChild; child != nil; child = child.NextSibling {
			if evaluator.match(matcher, child) {
				result = append(result, child)
			}
			visit(child)
		}
	}
	visit(root)
	return result
}

func queryWithEvaluator(root *html.Node, matcher Matcher, evaluator *queryEvaluator) *html.Node {
	for child := root.FirstChild; child != nil; child = child.NextSibling {
		if evaluator.match(matcher, child) {
			return child
		}
		if match := queryWithEvaluator(child, matcher, evaluator); match != nil {
			return match
		}
	}
	return nil
}
