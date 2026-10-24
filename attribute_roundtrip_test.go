package cascadia

import (
	"reflect"
	"testing"
)

func TestAttributeStringRoundTrip(t *testing.T) {
	for _, input := range []string{`[title='a"b']`, `[title="a\\b"]`, `[title="a\a b"]`, `[title="a\d b"]`, `[title="a\c b"]`, `[title="plain" i]`} {
		s, err := ParseGroupWithPseudoElements(input)
		if err != nil {
			t.Fatal(err)
		}
		got, err := ParseGroupWithPseudoElements(s.String())
		if err != nil || !reflect.DeepEqual(s, got) {
			t.Errorf("%q serialized as %q: %v", input, s.String(), err)
		}
	}
}
