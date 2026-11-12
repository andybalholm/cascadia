package cascadia

import (
	"reflect"
	"testing"
)

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
