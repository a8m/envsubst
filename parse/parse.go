// Most of the code in this package taken from golang/text/template/parse
package parse

import (
	"errors"
	"strings"
)

// A mode value is a set of flags (or 0). They control parser behavior.
type Mode int

// Mode for parser behaviour
const (
	Quick     Mode = iota // stop parsing after first error encountered and return
	AllErrors             // report all errors
)

// The restrictions option controls the parsing restriction.
type Restrictions struct {
	NoUnset bool
	NoEmpty bool
	NoDigit bool
}

// Restrictions specifier
var (
	Relaxed = &Restrictions{false, false, false}
	NoEmpty = &Restrictions{false, true, false}
	NoUnset = &Restrictions{true, false, false}
	Strict  = &Restrictions{true, true, false}
)

// Parser type initializer
type Parser struct {
	Name      string // name of the processing template
	Env       Env
	Restrict  *Restrictions
	Mode      Mode
	VarFilter *VarFilter // nil means all vars allowed; non-nil limits substitution
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
	p.lex = lex(text, p.Restrict.NoDigit)
	// Build internal array of all unset or empty vars here
	var errs []error
	// clean parse state
	p.nodes = make([]Node, 0)
	p.peekCount = 0
	if err := p.parse(); err != nil {
		switch p.Mode {
		case Quick:
			return "", err
		case AllErrors:
			errs = append(errs, err)
		}
	}
	var out string
	for _, node := range p.nodes {
		s, err := node.String()
		if err != nil {
			switch p.Mode {
			case Quick:
				return "", err
			case AllErrors:
				errs = append(errs, err)
			}
		}
		out += s
	}
	if len(errs) > 0 {
		var b strings.Builder
		for i, err := range errs {
			if i > 0 {
				b.WriteByte('\n')
			}
			b.WriteString(err.Error())
		}
		return "", errors.New(b.String())
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
			varNode := NewVariable(strings.TrimPrefix(t.val, "$"), p.Env, p.Restrict, p.VarFilter)
			varNode.OriginalSrc = t.val // Store original like "$VAR"
			p.nodes = append(p.nodes, varNode)
		case itemLeftDelim:
			if p.peek().typ == itemVariable {
				n, err := p.action(t.pos) // Pass the position of ${
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
// delimPos is the position of the opening ${
func (p *Parser) action(delimPos Pos) (Node, error) {
	var expType itemType
	var defaultNode Node
	varNode := NewVariable(p.next().val, p.Env, p.Restrict, p.VarFilter)
	for {
		switch t := p.next(); t.typ {
		case itemRightDelim:
			// Capture original source from ${ through }
			// t.pos is the position of }, and we need to include it
			endPos := int(t.pos) + len(t.val)
			if endPos > len(p.lex.input) {
				endPos = len(p.lex.input)
			}
			originalSrc := p.lex.input[delimPos:endPos]
			node := &SubstitutionNode{
				NodeType:    NodeSubstitution,
				ExpType:     expType,
				Variable:    varNode,
				Default:     defaultNode,
				OriginalSrc: originalSrc,
			}
			return node, nil
		case itemError:
			return nil, p.errorf(t.val)
		case itemVariable:
			defVar := NewVariable(strings.TrimPrefix(t.val, "$"), p.Env, p.Restrict, p.VarFilter)
			defVar.OriginalSrc = t.val
			defaultNode = defVar
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
