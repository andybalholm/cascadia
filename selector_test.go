package cascadia

import (
	"fmt"
	"html"
	"os"
	"strings"
	"testing"
)

type selectorTest struct {
	HTML, selector string
	testFunc func([]*html.Node) os.Error
}

var selectorTests = []selectorTest{
	{
		`<body><address>This address...</address></body>`,
		"address",
		func (r []*html.Node) os.Error {
			if len(r) != 1 {
				return fmt.Errorf("wanted one element, got %d", len(r))
			}

			if r[0].Data != "address" {
				return fmt.Errorf("wanted an address element, got %s", r[0].Data)
			}

			return nil
		},
	},
}

func TestSelectors(t *testing.T) {
	for _, test := range selectorTests {
		s, err := Compile(test.selector)
		if err != nil {
			t.Errorf("error compiling %q: %s", test.selector, err)
			continue
		}

		doc, err := html.Parse(strings.NewReader(test.HTML))
		if err != nil {
			t.Errorf("error parsing %q: %s", test.HTML, err)
			continue
		}

		err = test.testFunc(s.MatchAll(doc))
		if err != nil {
			t.Errorf("error in results of %q selector: %s", test.selector, err)
		}
	}
}
