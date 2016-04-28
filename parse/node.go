package parse

import (
	"os"
)

type Node interface {
	Type() NodeType
	String() string
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

func (t *TextNode) String() string {
	return t.Text
}

type VariableNode struct {
	NodeType
	Ident string
}

func NewVariable(ident string) *VariableNode {
	return &VariableNode{NodeVariable, ident}
}

func (t *VariableNode) String() string {
	return os.Getenv(t.Ident)
}

func (t *VariableNode) isSet() bool {
	_, isSet := os.LookupEnv(t.Ident)
	return isSet
}

type SubstitutionNode struct {
	NodeType
	ExpType  itemType
	Variable *VariableNode
	Default  Node // Default could be variable or text
}

func (t *SubstitutionNode) String() string {
	if t.ExpType >= itemPlus && t.Default != nil {
		switch t.ExpType {
		case itemColonDash, itemColonEquals:
			if s := t.Variable.String(); s != "" {
				return s
			}
			return t.Default.String()
		case itemPlus, itemColonPlus:
			if t.Variable.isSet() {
				return t.Default.String()
			}
		default:
			if !t.Variable.isSet() {
				return t.Default.String()
			}
		}
	}
	return t.Variable.String()
}
