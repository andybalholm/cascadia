package cascadia

import (
	"slices"
	"sort"

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
	sibling []siblingMetadata
}

type siblingMetadata struct {
	position, reversePosition         uint32
	typePosition, typeReversePosition uint32
	count, typeCount                  uint32
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
		count, typeCounts := countElementChildren(parent)
		var typePositions map[elementName]int
		if count > 0 {
			typePositions = make(map[elementName]int, len(typeCounts))
		}
		siblingPosition := 0
		for n := parent.FirstChild; n != nil; n = n.NextSibling {
			if n.Type == html.ElementNode {
				siblingPosition++
				name := elementName{namespace: n.Namespace, data: n.Data}
				typePositions[name]++
				typePosition := typePositions[name]
				d.order[n] = position
				position++
				d.sibling = append(d.sibling, siblingMetadata{
					position:            uint32(siblingPosition),
					reversePosition:     uint32(count - siblingPosition + 1),
					typePosition:        uint32(typePosition),
					typeReversePosition: uint32(typeCounts[name] - typePosition + 1),
					count:               uint32(count),
					typeCount:           uint32(typeCounts[name]),
				})
				d.byTag[n.Data] = append(d.byTag[n.Data], n)
				var indexedIDs []string
				var indexedClasses []string
				for _, attr := range n.Attr {
					switch attr.Key {
					case "id":
						if !slices.Contains(indexedIDs, attr.Val) {
							indexedIDs = append(indexedIDs, attr.Val)
							d.byID[attr.Val] = append(d.byID[attr.Val], n)
						}
					case "class":
						indexedClasses = d.indexClasses(n, attr.Val, indexedClasses)
					}
				}
			}
			visit(n)
		}
	}
	visit(root)
	return d
}

func countElementChildren(parent *html.Node) (int, map[elementName]int) {
	count := 0
	var typeCounts map[elementName]int
	for child := parent.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.ElementNode {
			continue
		}
		count++
		if typeCounts == nil {
			typeCounts = make(map[elementName]int)
		}
		typeCounts[elementName{namespace: child.Namespace, data: child.Data}]++
	}
	return count, typeCounts
}

func (d *IndexedDocument) indexClasses(n *html.Node, value string, indexed []string) []string {
	for start := 0; start < len(value); {
		for start < len(value) && isCSSWhitespace(value[start]) {
			start++
		}
		end := start
		for end < len(value) && !isCSSWhitespace(value[end]) {
			end++
		}
		if end > start {
			class := value[start:end]
			if !slices.Contains(indexed, class) {
				indexed = append(indexed, class)
				d.byClass[class] = append(d.byClass[class], n)
			}
		}
		start = end
	}
	return indexed
}

func isCSSWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '\f'
}

// Query returns the first descendant matching m.
func (d *IndexedDocument) Query(m Matcher) *html.Node {
	if d == nil || d.root == nil || m == nil {
		return nil
	}
	candidates, ok := d.candidates(m)
	if !ok {
		return queryWithEvaluator(d.root, m, newQueryEvaluator(d))
	}
	evaluator := newQueryEvaluator(d)
	for _, n := range candidates {
		if evaluator.match(m, n) {
			return n
		}
	}
	return nil
}

// QueryAll returns all descendants matching m in document order. Selectors
// returned by Parse or ParseGroup can use the indexes. Other Matcher
// implementations remain supported and fall back to a tree scan.
func (d *IndexedDocument) QueryAll(m Matcher) []*html.Node {
	return d.QueryAllInto(m, nil)
}

// QueryAllInto appends all descendants matching m to storage in document
// order and returns the resulting slice. Reusing storage reduces allocations in
// repeated-query workloads.
func (d *IndexedDocument) QueryAllInto(m Matcher, storage []*html.Node) []*html.Node {
	if d == nil || d.root == nil || m == nil {
		return storage
	}
	candidates, ok := d.candidates(m)
	if !ok {
		return queryAllWithEvaluatorInto(d.root, m, newQueryEvaluator(d), storage)
	}
	evaluator := newQueryEvaluator(d)
	for _, n := range candidates {
		if evaluator.match(m, n) {
			storage = append(storage, n)
		}
	}
	return storage
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
		return d.bestCandidates(group[0])
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

func (d *IndexedDocument) bestCandidates(sel Sel) ([]*html.Node, bool) {
	switch s := sel.(type) {
	case combinedSelector:
		if s.second != nil {
			return d.bestCandidates(s.second)
		}
		return d.bestCandidates(s.first)
	case compoundSelector:
		var best []*html.Node
		found := false
		for _, part := range s.selectors {
			var nodes []*html.Node
			matchedAnchor := true
			switch p := part.(type) {
			case idSelector:
				nodes = d.byID[p.id]
			case classSelector:
				nodes = d.byClass[p.class]
			case tagSelector:
				nodes = d.byTag[p.tag]
			default:
				matchedAnchor = false
			}
			if !matchedAnchor {
				continue
			}
			if !found || len(nodes) < len(best) {
				best = nodes
			}
			found = true
		}
		return best, found
	case idSelector:
		return d.byID[s.id], true
	case classSelector:
		return d.byClass[s.class], true
	case tagSelector:
		return d.byTag[s.tag], true
	default:
		return nil, false
	}
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
