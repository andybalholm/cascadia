package cascadia

import (
	"fmt"
	"strings"
	"testing"
)

func benchmarkDocument(sections, links int) string {
	var b strings.Builder
	b.WriteString("<main id=app>")
	for i := 0; i < sections; i++ {
		fmt.Fprintf(&b, `<section class="panel group-%d">`, i%10)
		for j := 0; j < links; j++ {
			fmt.Fprintf(&b, `<a class="link item-%d" href="/%d/%d">item</a>`, j%20, i, j)
		}
		b.WriteString("</section>")
	}
	b.WriteString("</main>")
	return b.String()
}

func BenchmarkQueryEngines(b *testing.B) {
	root := MustParseHTML(benchmarkDocument(200, 20))
	indexed := NewIndexedDocument(root)
	for _, query := range []string{
		"#app",
		".item-19",
		"section.group-9 > a.item-19",
		"main section a[href$=\"/19\"]",
		"section:has(a.item-19)",
	} {
		matcher := mustParseGroup(query)
		b.Run(query, func(b *testing.B) {
			b.Run("scan", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					_ = QueryAll(root, matcher)
				}
			})
			b.Run("indexed", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					_ = indexed.QueryAll(matcher)
				}
			})
		})
	}
}

func BenchmarkBuildIndexedDocument(b *testing.B) {
	root := MustParseHTML(benchmarkDocument(200, 20))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewIndexedDocument(root)
	}
}
