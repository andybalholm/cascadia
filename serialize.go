package cascadia

import (
	"fmt"
	"strings"
)

// implements the reverse operation Sel -> string

func (c tagSelector) String() string {
	return c.tag
}

func (c idSelector) String() string {
	return "#" + c.id
}

func (c classSelector) String() string {
	return "." + c.class
}

func (c attrSelector) String() string {
	val := c.val
	if c.operation == "#=" {
		val = c.regexp.String()
	}
	return fmt.Sprintf("[%s%s%s]", c.key, c.operation, val)
}

func (c relativePseudoClassSelector) String() string {
	panic("not yet supported")
}
func (c containsPseudoClassSelector) String() string {
	panic("not yet supported")
}
func (c regexpPseudoClassSelector) String() string {
	panic("not yet supported")
}
func (c nthPseudoClassSelector) String() string {
	panic("not yet supported")
}
func (c onlyChildPseudoClassSelector) String() string {
	panic("not yet supported")
}
func (c inputPseudoClassSelector) String() string {
	panic("not yet supported")
}
func (c emptyElementPseudoClassSelector) String() string {
	panic("not yet supported")
}
func (c rootPseudoClassSelector) String() string {
	panic("not yet supported")
}

func (c compoundSelector) String() string {
	chunks := make([]string, len(c.selectors))
	for i, sel := range c.selectors {
		chunks[i] = sel.String()
	}
	s := strings.Join(chunks, "")
	if c.pseudoElement != "" {
		s += "::" + c.pseudoElement
	}
	return s
}

func (c combinedSelector) String() string {
	start := c.first.String()
	if c.second != nil {
		start += fmt.Sprintf("%s %s %s", start, string(c.combinator), c.second.String())
	}
	return start
}
