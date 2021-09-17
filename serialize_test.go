package cascadia

import (
	"reflect"
	"testing"
)

var testSer []string

func init() {
	for _, test := range selectorTests {
		testSer = append(testSer, test.selector)
	}
	for _, test := range testsPseudo {
		testSer = append(testSer, test.selector)
	}
}

func TestSerialize(t *testing.T) {

	for _, test := range testSer {
		s, err := ParseGroupWithPseudoElements(test)
		if err != nil {
			t.Fatalf("error compiling %q: %s", test, err)
		}
		serialized := s.String()
		s2, err := ParseGroupWithPseudoElements(serialized)
		if err != nil {
			t.Fatalf("error compiling %q: %s %T (original : %s)", serialized, err, s, test)
		}

		if !reflect.DeepEqual(s, s2) {
			t.Fatalf("can't retrieve selector from serialized : %s (original : %s, sel : %#v)", serialized, test, s)
		}
	}
}
