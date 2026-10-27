package cascadia

import (
	"slices"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

// IndexedDocument accelerates repeated selector queries over an immutable HTML
// tree. The index must be rebuilt after the tree is mutated.
//
// IndexedDocument is most useful when many queries are executed against the
// same document. A single query may be faster with QueryAll because building
// the index has a cost of its own.
type IndexedDocument struct {
	root    *html.Node
	byID    map[string][]*html.Node
	byClass map[string][]*html.Node
	byTag   map[string][]*html.Node
	order   map[*html.Node]int
}

// NewIndexedDocument builds indexes for the descendants of root. This matches
// QueryAll's behavior, which does not test root itself.
func NewIndexedDocument(root *html.Node) *IndexedDocument {
	d := &IndexedDocument{
		root:    root,
		byID:    make(map[string][]*html.Node),
		byClass: make(map[string][]*html.Node),
		byTag:   make(map[string][]*html.Node),
		order:   make(map[*html.Node]int),
	}
	if root == nil {
		return d
	}

	position := 0
	var visit func(*html.Node)
	visit = func(parent *html.Node) {
		for n := parent.FirstChild; n != nil; n = n.NextSibling {
			if n.Type == html.ElementNode {
				d.order[n] = position
				position++
				d.byTag[n.Data] = append(d.byTag[n.Data], n)
				var indexedClasses []string
				for _, attr := range n.Attr {
					switch attr.Key {
					case "id":
						d.byID[attr.Val] = append(d.byID[attr.Val], n)
					case "class":
						for _, class := range strings.Fields(attr.Val) {
							if slices.Contains(indexedClasses, class) {
								continue
							}
							indexedClasses = append(indexedClasses, class)
							d.byClass[class] = append(d.byClass[class], n)
						}
					}
				}
			}
			visit(n)
		}
	}
	visit(root)
	return d
}

// Query returns the first descendant matching m.
func (d *IndexedDocument) Query(m Matcher) *html.Node {
	if d == nil || d.root == nil || m == nil {
		return nil
	}
	candidates, ok := d.candidates(m)
	if !ok {
		return Query(d.root, m)
	}
	for _, n := range candidates {
		if m.Match(n) {
			return n
		}
	}
	return nil
}

// QueryAll returns all descendants matching m in document order. Selectors
// returned by Parse or ParseGroup can use the indexes. Other Matcher
// implementations remain supported and fall back to a tree scan.
func (d *IndexedDocument) QueryAll(m Matcher) []*html.Node {
	if d == nil || d.root == nil || m == nil {
		return nil
	}
	candidates, ok := d.candidates(m)
	if !ok {
		return QueryAll(d.root, m)
	}
	result := make([]*html.Node, 0, len(candidates))
	for _, n := range candidates {
		if m.Match(n) {
			result = append(result, n)
		}
	}
	return result
}

type indexAnchor struct {
	kind  byte
	value string
}

func (d *IndexedDocument) candidates(m Matcher) ([]*html.Node, bool) {
	group, ok := m.(SelectorGroup)
	if !ok {
		sel, ok := m.(Sel)
		if !ok {
			return nil, false
		}
		group = SelectorGroup{sel}
	}
	if len(group) == 1 {
		anchors := rightmostAnchors(group[0])
		if len(anchors) == 0 {
			return nil, false
		}
		best := d.nodesForAnchor(anchors[0])
		for _, anchor := range anchors[1:] {
			if nodes := d.nodesForAnchor(anchor); len(nodes) < len(best) {
				best = nodes
			}
		}
		return best, true
	}

	seen := make(map[*html.Node]struct{})
	var candidates []*html.Node
	for _, sel := range group {
		anchors := rightmostAnchors(sel)
		if len(anchors) == 0 {
			return nil, false
		}
		best := d.nodesForAnchor(anchors[0])
		for _, anchor := range anchors[1:] {
			if nodes := d.nodesForAnchor(anchor); len(nodes) < len(best) {
				best = nodes
			}
		}
		for _, n := range best {
			if _, exists := seen[n]; !exists {
				seen[n] = struct{}{}
				candidates = append(candidates, n)
			}
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return d.order[candidates[i]] < d.order[candidates[j]]
	})
	return candidates, true
}

func (d *IndexedDocument) nodesForAnchor(anchor indexAnchor) []*html.Node {
	switch anchor.kind {
	case '#':
		return d.byID[anchor.value]
	case '.':
		return d.byClass[anchor.value]
	default:
		return d.byTag[anchor.value]
	}
}

func rightmostAnchors(sel Sel) []indexAnchor {
	switch s := sel.(type) {
	case combinedSelector:
		if s.second == nil {
			return rightmostAnchors(s.first)
		}
		return rightmostAnchors(s.second)
	case compoundSelector:
		var anchors []indexAnchor
		for _, part := range s.selectors {
			switch p := part.(type) {
			case idSelector:
				anchors = append(anchors, indexAnchor{kind: '#', value: p.id})
			case classSelector:
				anchors = append(anchors, indexAnchor{kind: '.', value: p.class})
			case tagSelector:
				anchors = append(anchors, indexAnchor{kind: 't', value: p.tag})
			}
		}
		return anchors
	case idSelector:
		return []indexAnchor{{kind: '#', value: s.id}}
	case classSelector:
		return []indexAnchor{{kind: '.', value: s.class}}
	case tagSelector:
		return []indexAnchor{{kind: 't', value: s.tag}}
	default:
		return nil
	}
}
