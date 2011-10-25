// The cascadia package is an implementation of CSS selectors.
package cascadia

import (
	"fmt"
	"html"
	"os"
	"strconv"
)

// a parser for CSS selectors
type parser struct {
	s string // the source text
	i int    // the current position
}

// parseEscape parses a backslash escape.
func (p *parser) parseEscape() (result string, err os.Error) {
	if len(p.s) < p.i+2 || p.s[p.i] != '\\' {
		return "", os.NewError("invalid escape sequence")
	}

	start := p.i + 1
	c := p.s[start]
	switch {
	case c == '\r' || c == '\n' || c == '\f':
		return "", os.NewError("escaped line ending outside string")
	case hexDigit(c):
		// unicode escape (hex)
		var i int
		for i = start; i < p.i+6 && i < len(p.s) && hexDigit(p.s[i]); i++ {
			// empty
		}
		v, _ := strconv.Btoui64(p.s[start:i], 16)
		if len(p.s) > i {
			switch p.s[i] {
			case '\r':
				i++
				if len(p.s) > i && p.s[i] == '\n' {
					i++
				}
			case ' ', '\t', '\n', '\f':
				i++
			}
		}
		p.i = i
		return string(int(v)), nil
	}

	// Return the literal character after the backslash.
	result = p.s[start : start+1]
	p.i += 2
	return result, nil
}

func hexDigit(c byte) bool {
	return '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F'
}

// nameStart returns whether c can be the first character of an identifier
// (not counting an initial hyphen, or an escape sequence).
func nameStart(c byte) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_' || c > 127
}

// nameChar returns whether c can be a character within an identifier
// (not counting an escape sequence).
func nameChar(c byte) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_' || c > 127 ||
		c == '-' || '0' <= c && c <= '9'
}

// parseIdentifier parses an identifier.
func (p *parser) parseIdentifier() (result string, err os.Error) {
	startingDash := false
	if len(p.s) > p.i && p.s[p.i] == '-' {
		startingDash = true
		p.i++
	}

	if len(p.s) <= p.i {
		return "", os.NewError("expected identifier, found EOF instead")
	}

	if c := p.s[p.i]; !(nameStart(c) || c == '\\') {
		return "", fmt.Errorf("expected identifier, found %c instead", c)
	}

	result, err = p.parseName()
	if startingDash && err == nil {
		result = "-" + result
	}
	return
}

// parseName parses a name (which is like an identifier, but doesn't have
// extra restrictions on the first character).
func (p *parser) parseName() (result string, err os.Error) {
	i := p.i
loop:
	for i < len(p.s) {
		c := p.s[i]
		switch {
		case nameChar(c):
			start := i
			for i < len(p.s) && nameChar(p.s[i]) {
				i++
			}
			result += p.s[start:i]
		case c == '\\':
			p.i = i
			val, err := p.parseEscape()
			if err != nil {
				return "", err
			}
			i = p.i
			result += val
		default:
			break loop
		}
	}

	if result == "" {
		return "", os.NewError("expected name, found EOF instead")
	}

	p.i = i
	return result, nil
}

// parseString parses a single- or double-quoted string.
func (p *parser) parseString() (result string, err os.Error) {
	i := p.i
	if len(p.s) < i+2 {
		return "", os.NewError("expected string, found EOF instead")
	}

	quote := p.s[i]
	i++

loop:
	for i < len(p.s) {
		switch p.s[i] {
		case '\\':
			if len(p.s) > i+1 {
				switch c := p.s[i+1]; c {
				case '\r':
					if len(p.s) > i+2 && p.s[i+2] == '\n' {
						i += 3
						continue loop
					}
					fallthrough
				case '\n', '\f':
					i += 2
					continue loop
				}
			}
			p.i = i
			val, err := p.parseEscape()
			if err != nil {
				return "", err
			}
			i = p.i
			result += val
		case quote:
			break loop
		case '\r', '\n', '\f':
			return "", os.NewError("unexpected end of line in string")
		default:
			start := i
			for i < len(p.s) {
				if c := p.s[i]; c == quote || c == '\\' || c == '\r' || c == '\n' || c == '\f' {
					break
				}
				i++
			}
			result += p.s[start:i]
		}
	}

	if i >= len(p.s) {
		return "", os.NewError("EOF in string")
	}

	// Consume the final quote.
	i++

	p.i = i
	return result, nil
}

// parseTypeSelector parses a type selector (one that matches by tag name).
func (p *parser) parseTypeSelector() (result Selector, err os.Error) {
	tag, err := p.parseIdentifier()
	if err != nil {
		return nil, err
	}

	return typeSelector(tag), nil
}

// parseIDSelector parses a selector that matches by id attribute.
func (p *parser) parseIDSelector() (Selector, os.Error) {
	if p.i >= len(p.s) {
		return nil, fmt.Errorf("expected id selector (#id), found EOF instead")
	}
	if p.s[p.i] != '#' {
		return nil, fmt.Errorf("expected id selector (#id), found '%c' instead", p.s[p.i])
	}

	p.i++
	id, err := p.parseName()
	if err != nil {
		return nil, err
	}

	return attributeEqualsSelector("id", id), nil
}

// parseSimpleSelectorSequence parses a selector sequence that applies to
// a single element.
func (p *parser) parseSimpleSelectorSequence() (Selector, os.Error) {
	var result Selector

	if p.i >= len(p.s) {
		return nil, os.NewError("expected selector, found EOF instead")
	}

	switch p.s[p.i] {
	case '*':
		// It's the universal selector. Just skip over it, since it doesn't affect the meaning.
		p.i++
	case '#', '.', '[', ':':
		// There's no type selector. Wait to process the other till the main loop.
	default:
		r, err := p.parseTypeSelector()
		if err != nil {
			return nil, err
		}
		result = r
	}

loop:
	for p.i < len(p.s) {
		var ns Selector
		var err os.Error
		switch p.s[p.i] {
		case '#':
			ns, err = p.parseIDSelector()
		case '.':
			// TODO: parseClassSelector
		case '[':
			// TODO: parseAttributeSelector
		case ':':
			// TODO: parsePseudoclassSelector
		default:
			break loop
		}
		if err != nil {
			return nil, err
		}
		if result == nil {
			result = ns
		} else {
			result = intersectionSelector(result, ns)
		}
	}

	if result == nil {
		result = func(n *html.Node) bool {
			return true
		}
	}

	return result, nil
}
