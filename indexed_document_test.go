package cascadia

import (
	"reflect"
	"testing"

	"golang.org/x/net/html"
)

func TestIndexedDocumentMatchesQueryAll(t *testing.T) {
	root := MustParseHTML(`
		<main id="app">
			<section class="panel featured"><a class="link" href="/one">one</a></section>
			<section class="panel"><a class="link active" href="/two">two</a></section>
			<aside><a class="link" href="/three">three</a></aside>
		</main>`)
	doc := NewIndexedDocument(root)

	selectors := []string{
		"#app", ".panel", "section", "section.panel.featured",
		"main > section .link", "#app .active", "aside a, section.featured a",
		"[href]", ":first-child", "section:has(.active)",
	}
	for _, query := range selectors {
		t.Run(query, func(t *testing.T) {
			matcher := mustParseGroup(query)
			want := QueryAll(root, matcher)
			got := doc.QueryAll(matcher)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("QueryAll mismatch: got %p, want %p", got, want)
			}
			if first := doc.Query(matcher); first != firstNode(want) {
				t.Fatalf("Query mismatch: got %p, want %p", first, firstNode(want))
			}
		})
	}
}

func TestIndexedDocumentNilSafety(t *testing.T) {
	var doc *IndexedDocument
	if got := doc.QueryAll(mustParseGroup("div")); got != nil {
		t.Fatalf("nil document returned %#v", got)
	}
	if got := NewIndexedDocument(nil).Query(mustParseGroup("div")); got != nil {
		t.Fatalf("nil root returned %p", got)
	}
}

func TestIndexedDocumentRequiresRebuildAfterMutation(t *testing.T) {
	root := MustParseHTML(`<main><p class="old"></p></main>`)
	doc := NewIndexedDocument(root)
	p := Query(root, mustParseGroup("p"))
	p.Attr[0].Val = "new"

	if got := doc.QueryAll(mustParseGroup(".new")); len(got) != 0 {
		t.Fatalf("stale index unexpectedly observed mutation")
	}
	if got := NewIndexedDocument(root).QueryAll(mustParseGroup(".new")); len(got) != 1 {
		t.Fatalf("rebuilt index returned %d matches", len(got))
	}
}

func TestIndexedDocumentDoesNotDuplicateRepeatedClasses(t *testing.T) {
	root := MustParseHTML(`<main><p class="item item"></p></main>`)
	got := NewIndexedDocument(root).QueryAll(mustParseGroup(".item"))
	if len(got) != 1 {
		t.Fatalf("returned %d matches for one element", len(got))
	}
}

func TestIndexedDocumentDoesNotDuplicateRepeatedIDs(t *testing.T) {
	root := MustParseHTML(`<main><p></p></main>`)
	p := Query(root, mustParseGroup("p"))
	p.Attr = []html.Attribute{
		{Key: "id", Val: "item"},
		{Key: "id", Val: "item"},
	}

	got := NewIndexedDocument(root).QueryAll(mustParseGroup("#item"))
	if len(got) != 1 {
		t.Fatalf("returned %d matches for one element", len(got))
	}
}

func mustParseGroup(query string) SelectorGroup {
	matcher, err := ParseGroup(query)
	if err != nil {
		panic(err)
	}
	return matcher
}

func firstNode(nodes []*html.Node) *html.Node {
	if len(nodes) == 0 {
		return nil
	}
	return nodes[0]
}
