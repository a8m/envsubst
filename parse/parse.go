// Most of the code in this package taken from golang/text/template/parse
package parse

import (
	"errors"
	"strings"
)

type Restrictions struct {
	NoUnset bool
	NoEmpty bool
}

var Relaxed = &Restrictions{false, false}
var NoEmpty = &Restrictions{false, true}
var NoUnset = &Restrictions{true, false}
var Strict = &Restrictions{true, true}

type Parser struct {
	Name     string // name of the processing template
	Env      Env
	Restrict *Restrictions
	// parsing state;
	lex       *lexer
	token     [3]item // three-token lookahead
	peekCount int
	nodes     []Node
}

// New allocates a new Parser with the given name.
func New(name string, env []string, r *Restrictions) *Parser {
	return &Parser{
		Name:     name,
		Env:      Env(env),
		Restrict: r,
	}
}

// Parse parses the given string.
func (p *Parser) Parse(text string) (string, error) {
	p.lex = lex(text)
	// clean parse state
	p.nodes = make([]Node, 0)
	p.peekCount = 0
	if err := p.parse(); err != nil {
		return "", err
	}
	var out string
	for _, node := range p.nodes {
		s, err := node.String()
		if err != nil {
			return out, err
		}
		out += s
	}
	return out, nil
}

// parse is the top-level parser for the template.
// It runs to EOF and return an error if something isn't right.
func (p *Parser) parse() error {
Loop:
	for {
		switch t := p.next(); t.typ {
		case itemEOF:
			break Loop
		case itemError:
			return p.errorf(t.val)
		case itemVariable:
			varNode := NewVariable(strings.TrimPrefix(t.val, "$"), p.Env, p.Restrict)
			p.nodes = append(p.nodes, varNode)
		case itemLeftDelim:
			if p.peek().typ == itemVariable {
				n, err := p.action()
				if err != nil {
					return err
				}
				p.nodes = append(p.nodes, n)
				continue
			}
			fallthrough
		default:
			textNode := NewText(t.val)
			p.nodes = append(p.nodes, textNode)
		}
	}
	return nil
}

// Parse substitution. first item is a variable.
func (p *Parser) action() (Node, error) {
	var expType itemType
	var defaultNode Node
	varNode := NewVariable(p.next().val, p.Env, p.Restrict)
Loop:
	for {
		switch t := p.next(); t.typ {
		case itemRightDelim:
			break Loop
		case itemError:
			return nil, p.errorf(t.val)
		case itemVariable:
			defaultNode = NewVariable(strings.TrimPrefix(t.val, "$"), p.Env, p.Restrict)
		case itemText:
			n := NewText(t.val)
		Text:
			for {
				switch p.peek().typ {
				case itemRightDelim, itemError, itemEOF:
					break Text
				default:
					// patch to accept all kind of chars
					n.Text += p.next().val
				}
			}
			defaultNode = n
		default:
			expType = t.typ
		}
	}
	return &SubstitutionNode{NodeSubstitution, expType, varNode, defaultNode}, nil
}

func (p *Parser) errorf(s string) error {
	return errors.New(s)
}

// next returns the next token.
func (p *Parser) next() item {
	if p.peekCount > 0 {
		p.peekCount--
	} else {
		p.token[0] = p.lex.nextItem()
	}
	return p.token[p.peekCount]
}

// backup backs the input stream up one token.
func (p *Parser) backup() {
	p.peekCount++
}

// peek returns but does not consume the next token.
func (p *Parser) peek() item {
	if p.peekCount > 0 {
		return p.token[p.peekCount-1]
	}
	p.peekCount = 1
	p.token[0] = p.lex.nextItem()
	return p.token[0]
}
