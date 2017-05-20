package parse

import (
	"fmt"
)

type Node interface {
	Type() NodeType
	String() (string, error)
}

// NodeType identifies the type of a node.
type NodeType int

// Type returns itself and provides an easy default implementation
// for embedding in a Node. Embedded in all non-trivial Nodes.
func (t NodeType) Type() NodeType {
	return t
}

const (
	NodeText NodeType = iota
	NodeSubstitution
	NodeVariable
)

type TextNode struct {
	NodeType
	Text string
}

func NewText(text string) *TextNode {
	return &TextNode{NodeText, text}
}

func (t *TextNode) String() (string, error) {
	return t.Text, nil
}

type VariableNode struct {
	NodeType
	Ident string
	Env   Env
}

func NewVariable(ident string, env Env) *VariableNode {
	return &VariableNode{NodeVariable, ident, env}
}

func (t *VariableNode) String() (string, error) {
	return t.Env.Get(t.Ident), nil
}

func (t *VariableNode) isSet() bool {
	return t.Env.Has(t.Ident)
}

type SubstitutionNode struct {
	NodeType
	ExpType  itemType
	Variable *VariableNode
	Default  Node // Default could be variable or text
	Restrict *Restrictions
}

func (t *SubstitutionNode) String() (string, error) {
	if t.ExpType >= itemPlus && t.Default != nil {
		switch t.ExpType {
		case itemColonDash, itemColonEquals:
			if s, _ := t.Variable.String(); s != "" {
				return s, nil
			}
			return t.validate(t.Default)
		case itemPlus, itemColonPlus:
			if t.Variable.isSet() {
				return t.validate(t.Default)
			}
			return "", nil
		default:
			if !t.Variable.isSet() {
				return t.validate(t.Default)
			}
		}
	}
	return t.validate(t.Variable)
}

func (t *SubstitutionNode) validate(node Node) (string, error) {
	if err := t.validateNoUnset(node); err != nil {
		return "", err
	}
	return t.validateNoEmpty(node)
}

func (t *SubstitutionNode) validateNoUnset(node Node) error {
	if t.Restrict.NoUnset && node.Type() == NodeVariable && !node.(*VariableNode).isSet() {
		return fmt.Errorf("variable ${%s} not set", t.Variable.Ident)
	}
	return nil
}

func (t *SubstitutionNode) validateNoEmpty(node Node) (string, error) {
	value, _ := node.String()
	if t.Restrict.NoEmpty && value == "" && (node.Type() != NodeVariable || node.(*VariableNode).isSet()) {
		return "", fmt.Errorf("variable ${%s} set but empty", t.Variable.Ident)
	}
	return value, nil
}
