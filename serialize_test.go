package cascadia

import (
	"reflect"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestSerializeNumericClassAndIDNames(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(`<main>
<div id="0" class="0"></div>
<div id="123" class="123"></div>
<div id="2xl:flex" class="2xl:flex"></div>
<div id="9abc" class="9abc"></div>
<div id="-2xl" class="-2xl"></div>
<div id="-" class="-"></div>
<div id="--label" class="--label"></div>
<div id="alpha1" class="alpha1"></div>
<div id="éclair" class="éclair"></div>
</main>`))
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct{ input, want string }{
		{`\30 `, `\30 `},
		{`\31 23`, `\31 23`},
		{`\32 xl\:flex`, `\32 xl\:flex`},
		{`\39 abc`, `\39 abc`},
		{`\-2xl`, `\-2xl`},
		{`\-`, `\-`},
		{`--label`, `\-\-label`},
		{`alpha1`, `alpha1`},
		{`éclair`, `éclair`},
	} {
		for _, prefix := range []string{".", "#"} {
			t.Run(prefix+test.input, func(t *testing.T) {
				selector, err := Parse(prefix + test.input)
				if err != nil {
					t.Fatal(err)
				}
				before := QueryAll(doc, selector)
				if len(before) != 1 {
					t.Fatalf("original selector matched %d elements, want 1", len(before))
				}

				serialized := selector.String()
				if want := prefix + test.want; serialized != want {
					t.Errorf("String() = %q, want %q", serialized, want)
				}
				roundtrip, err := Parse(serialized)
				if err != nil {
					t.Fatalf("parsing serialized selector %q: %v", serialized, err)
				}
				after := QueryAll(doc, roundtrip)
				if len(after) != 1 || after[0] != before[0] {
					t.Errorf("serialized selector %q changed the matched elements", serialized)
				}
			})
		}
	}
}

func TestSerializeEscapesPseudoClassArguments(t *testing.T) {
	// The argument of :contains is written into a quoted string, so the
	// quotes and the backslashes in it have to be escaped.
	for _, selector := range []string{
		`:contains("a\"b")`,
		`:contains("a\\b")`,
		`:contains("a\"b\"c")`,
		`:containsOwn("a\"b")`,
	} {
		s, err := Parse(selector)
		if err != nil {
			t.Fatalf("error compiling %q: %s", selector, err)
		}

		serialized := s.String()
		s2, err := Parse(serialized)
		if err != nil {
			t.Errorf("error compiling %q: %s (original : %s)", serialized, err, selector)
			continue
		}
		if !reflect.DeepEqual(s, s2) {
			t.Errorf("can't retrieve selector from serialized : %s (original : %s)", serialized, selector)
		}
	}
}

func TestSerialize(t *testing.T) {
	var testSer []string
	for _, test := range selectorTests {
		testSer = append(testSer, test.selector)
	}
	for _, test := range testsPseudo {
		testSer = append(testSer, test.selector)
	}
	for _, test := range loadValidSelectors(t) {
		if test.Xfail {
			continue
		}
		testSer = append(testSer, test.Selector)
	}

	for _, test := range testSer {
		s, err := ParseGroupWithPseudoElements(test)
		if err != nil {
			t.Fatalf("error compiling %q: %s", test, err)
		}

		serialized := s.String()
		s2, err := ParseGroupWithPseudoElements(serialized)
		if err != nil {
			t.Errorf("error compiling %q: %s %T (original : %s)", serialized, err, s, test)
		}

		if !reflect.DeepEqual(s, s2) {
			t.Errorf("can't retrieve selector from serialized : %s (original : %s, sel : %#v)", serialized, test, s)
		}
	}
}
