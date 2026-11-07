package cascadia

import (
	"reflect"
	"sync"
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
		"[href]", ":first-child", ":last-child", ":nth-child(2)",
		":nth-of-type(2n+1)", ":only-child", "section:not(.featured)",
		"section:has(.active)", "section:has(> a.active)",
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

func TestQueryAllIntoAppendsAndReusesStorage(t *testing.T) {
	root := MustParseHTML(`<main><p class="item"></p><p class="item"></p></main>`)
	matcher := mustParseGroup(".item")
	sentinel := root

	for name, query := range map[string]func([]*html.Node) []*html.Node{
		"scan": func(storage []*html.Node) []*html.Node {
			return QueryAllInto(root, matcher, storage)
		},
		"indexed": func(storage []*html.Node) []*html.Node {
			return NewIndexedDocument(root).QueryAllInto(matcher, storage)
		},
	} {
		t.Run(name, func(t *testing.T) {
			storage := make([]*html.Node, 1, 4)
			storage[0] = sentinel
			got := query(storage)
			if len(got) != 3 || got[0] != sentinel {
				t.Fatalf("got %d nodes with prefix %p", len(got), got[0])
			}
			if &got[0] != &storage[0] {
				t.Fatal("query did not reuse available storage capacity")
			}
		})
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

func TestIndexedDocumentConcurrentQueries(t *testing.T) {
	root := MustParseHTML(benchmarkDocument(20, 10))
	doc := NewIndexedDocument(root)
	selectors := []SelectorGroup{
		mustParseGroup(".item-9"),
		mustParseGroup("a:nth-of-type(2n+1)"),
		mustParseGroup("section:has(> a.item-9)"),
	}

	var wait sync.WaitGroup
	for _, selector := range selectors {
		selector := selector
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range 50 {
				got := doc.QueryAll(selector)
				want := QueryAll(root, selector)
				if !reflect.DeepEqual(got, want) {
					t.Errorf("concurrent query mismatch for %s", selector)
					return
				}
			}
		}()
	}
	wait.Wait()
}

func TestIndexedDocumentRelativeSelectorOutsideIndexedSubtree(t *testing.T) {
	root := MustParseHTML(`<body><header></header><main><p></p></main></body>`)
	main := Query(root, mustParseGroup("main"))
	matcher := mustParseGroup("body:has(> header) p")

	got := NewIndexedDocument(main).QueryAll(matcher)
	want := QueryAll(main, matcher)
	if !reflect.DeepEqual(got, want) || len(got) != 1 {
		t.Fatalf("subtree query mismatch: got %d nodes, want %d", len(got), len(want))
	}
}

func FuzzIndexedDocumentMatchesQueryAll(f *testing.F) {
	root := MustParseHTML(benchmarkDocument(8, 6))
	doc := NewIndexedDocument(root)
	for _, selector := range []string{
		"#app", ".item-5", "section > a", "a:nth-child(2n+1)",
		"section:has(> a.item-5)", "section:not(.group-2)",
		"main a[href$=\"/5\"]", "section, a.item-3",
	} {
		f.Add(selector)
	}
	f.Fuzz(func(t *testing.T, selector string) {
		matcher, err := ParseGroup(selector)
		if err != nil {
			t.Skip()
		}
		got := doc.QueryAll(matcher)
		want := QueryAll(root, matcher)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("QueryAll mismatch for %q", selector)
		}
	})
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
