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
	Quick     Mode = iota // stop parsing after first error encoutered and return
	AllErrors             // report all errors
)

// The restrictions option controls the parsring restriction.
type Restrictions struct {
	NoUnset bool
	NoEmpty bool
}

// Restrictions specifier
var (
	Relaxed = &Restrictions{false, false}
	NoEmpty = &Restrictions{false, true}
	NoUnset = &Restrictions{true, false}
	Strict  = &Restrictions{true, true}
)

// Parser type initializer
type Parser struct {
	Name         string // name of the processing template
	Env          Env
	SelectedEnvs []string
	Restrict     *Restrictions
	Mode         Mode
	// parsing state;
	lex       *lexer
	token     [3]item // three-token lookahead
	peekCount int
	nodes     []Node
}

// New allocates a new Parser with the given name.
func New(name string, env []string, r *Restrictions, selectedEnvs []string) *Parser {
	return &Parser{
		Name:         name,
		Env:          Env(env),
		Restrict:     r,
		SelectedEnvs: selectedEnvs,
	}
}

// Parse parses the given string.
func (p *Parser) Parse(text string) (string, error) {
	p.lex = lex(text)
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

func isVarLookupable(value string, selectedEnvs []string) bool {
	lookupable := true
	if len(selectedEnvs) > 0 {
		lookupable = false
		for _, env := range selectedEnvs {
			if env == value {
				lookupable = true
				break
			}
		}
	}
	return lookupable
}

// parse is the top-level parser for the template.
// It runs to EOF and return an error if something isn't right.
func (p *Parser) parse() error {
	currentVariable := []string{}
Loop:
	for {
		switch t := p.next(); t.typ {
		case itemEOF:
			break Loop
		case itemError:
			return p.errorf(t.val)
		case itemVariable:
			if isVarLookupable(strings.TrimPrefix(t.val, "$"), p.SelectedEnvs) {
				varNode := NewVariable(strings.TrimPrefix(t.val, "$"), p.Env, p.Restrict)
				p.nodes = append(p.nodes, varNode)
			}
		case itemLeftDelim:
			currentVariable = append(currentVariable, t.val)
			p_peek := p.peek()
			currentVariable = append(currentVariable, p_peek.val)
			if p_peek.typ == itemVariable {
				node, rightdelim, err := p.action()
				if err != nil {
					return err
				}
				if !isVarLookupable(p_peek.val, p.SelectedEnvs) {
					currentVariable = append(currentVariable, rightdelim)
					p.nodes = append(p.nodes, NewText(strings.Join(currentVariable, "")))
				} else {
					p.nodes = append(p.nodes, node)
				}
				currentVariable = nil
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
func (p *Parser) action() (Node, string, error) {
	var expType itemType
	var defaultNode Node
	var rightDelim string
	next_t_val := p.next().val
	varNode := NewVariable(next_t_val, p.Env, p.Restrict)

Loop:
	for {
		switch t := p.next(); t.typ {
		case itemRightDelim:
			rightDelim = t.val
			break Loop
		case itemError:
			return nil, rightDelim, p.errorf(t.val)
		case itemVariable:
			if isVarLookupable(strings.TrimPrefix(t.val, "$"), p.SelectedEnvs) {
				defaultNode = NewVariable(strings.TrimPrefix(t.val, "$"), p.Env, p.Restrict)
			}
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
	return &SubstitutionNode{NodeSubstitution, expType, varNode, defaultNode}, rightDelim, nil

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
