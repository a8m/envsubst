package parse

import (
	"testing"
)

func Test_lex(t *testing.T) {
	type args struct {
		input   string
		noDigit bool
	}
	tests := []struct {
		name string
		args args
		want []item
	}{
		{"empty", args{input: ""}, []item{tEOF}},
		{"text", args{input: "hello"}, []item{
			tText("hello"),
			tEOF,
		}},
		{"var", args{input: "$hello"}, []item{
			tVariable("$hello"),
			tEOF,
		}},
		{"single char var", args{input: "${A}"}, []item{
			tLeft,
			tVariable("A"),
			tRight,
			tEOF,
		}},
		{"2 vars", args{input: "$hello $world"}, []item{
			tVariable("$hello"),
			tText(" "),
			tVariable("$world"),
			tEOF,
		}},
		{"substitution-1", args{input: "bar ${BAR}"}, []item{
			tText("bar "),
			tLeft,
			tVariable("BAR"),
			tRight,
			tEOF,
		}},
		{"substitution-2", args{input: "bar ${BAR:=baz}"}, []item{
			tText("bar "),
			tLeft,
			tVariable("BAR"),
			tColEquals,
			tText("b"),
			tText("a"),
			tText("z"),
			tRight,
			tEOF,
		}},
		{"substitution-3", args{input: "bar ${BAR:=$BAZ}"}, []item{
			tText("bar "),
			tLeft,
			tVariable("BAR"),
			tColEquals,
			tVariable("$BAZ"),
			tRight,
			tEOF,
		}},
		{"substitution-4", args{input: "bar ${BAR:=$BAZ} foo"}, []item{
			tText("bar "),
			tLeft,
			tVariable("BAR"),
			tColEquals,
			tVariable("$BAZ"),
			tRight,
			tText(" foo"),
			tEOF,
		}},
		{"substitution-plus", args{input: "bar ${BAR+$BAZ} foo"}, []item{
			tText("bar "),
			tLeft,
			tVariable("BAR"),
			tPlus,
			tVariable("$BAZ"),
			tRight,
			tText(" foo"),
			tEOF,
		}},
		{"substitution-dash", args{input: "bar ${BAR-$BAZ} foo"}, []item{
			tText("bar "),
			tLeft,
			tVariable("BAR"),
			tDash,
			tVariable("$BAZ"),
			tRight,
			tText(" foo"),
			tEOF,
		}},
		{"substitution-equals", args{input: "bar ${BAR=$BAZ} foo"}, []item{
			tText("bar "),
			tLeft,
			tVariable("BAR"),
			tEquals,
			tVariable("$BAZ"),
			tRight,
			tText(" foo"),
			tEOF,
		}},
		{"substitution-col-plus", args{input: "bar ${BAR:+$BAZ} foo"}, []item{
			tText("bar "),
			tLeft,
			tVariable("BAR"),
			tColPlus,
			tVariable("$BAZ"),
			tRight,
			tText(" foo"),
			tEOF,
		}},
		{"substitution-leading-dash-1", args{input: "bar ${BAR:--1} foo"}, []item{
			tText("bar "),
			tLeft,
			tVariable("BAR"),
			tColDash,
			tText("-"),
			tText("1"),
			tRight,
			tText(" foo"),
			tEOF,
		}},
		{"substitution-leading-dash-2", args{input: "bar ${BAR:=-1} foo"}, []item{
			tText("bar "),
			tLeft,
			tVariable("BAR"),
			tColEquals,
			tText("-"),
			tText("1"),
			tRight,
			tText(" foo"),
			tEOF,
		}},
		{"closing brace error", args{input: "hello-${world"}, []item{
			tText("hello-"),
			tLeft,
			tVariable("world"),
			tError("closing brace expected"),
		}},
		{"closing brace error after default", args{input: "hello-${world:=1"}, []item{
			tText("hello-"),
			tLeft,
			tVariable("world"),
			tColEquals,
			tText("1"),
			tError("closing brace expected"),
		}},
		{"escaping $$var", args{input: "hello $$HOME"}, []item{
			tText("hello "),
			tText("$"),
			tText("HOME"),
			tEOF,
		}},
		{"escaping $${subst}", args{input: "hello $${HOME}"}, []item{
			tText("hello "),
			tText("$"),
			tText("{HOME}"),
			tEOF,
		}},
		{"starting with underscore 1", args{input: "hello $_"}, []item{
			tText("hello "),
			tText("$_"),
			tEOF,
		}},
		{"starting with underscore 2", args{input: "hello ${_}"}, []item{
			tText("hello "),
			tLeft,
			tText("_}"),
			tEOF,
		}},
		{"no digit $1", args{input: "hello $1", noDigit: true}, []item{
			tText("hello "),
			tText("$1"),
			tEOF,
		}},
		{"no digit $1ABC", args{input: "hello $1ABC", noDigit: true}, []item{
			tText("hello "),
			tText("$1"),
			tText("ABC"),
			tEOF,
		}},
		{"no digit ${2}", args{input: "hello ${2}", noDigit: true}, []item{
			tText("hello "),
			tText("${2"),
			tText("}"),
			tEOF,
		}},
		{"no digit ${2ABC}", args{input: "hello ${2ABC}", noDigit: true}, []item{
			tText("hello "),
			tText("${2"),
			tText("ABC}"),
			tEOF,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lex(tt.args.input, tt.args.noDigit)
			// gather the emitted items into "got"
			var got []item
			for {
				i := l.nextItem()
				got = append(got, i)
				if i.typ == itemEOF || i.typ == itemError {
					break
				}
			}
			if !equal(tt.want, got) {
				t.Errorf("lex(\"%s\") = %v, want %v", tt.args.input, got, tt.want)
			}
		})
	}
}

var (
	tEOF       = item{itemEOF, 0, ""}
	tPlus      = item{itemPlus, 0, "+"}
	tDash      = item{itemDash, 0, "-"}
	tEquals    = item{itemEquals, 0, "="}
	tColEquals = item{itemColonEquals, 0, ":="}
	tColDash   = item{itemColonDash, 0, ":-"}
	tColPlus   = item{itemColonPlus, 0, ":+"}
	tLeft      = item{itemLeftDelim, 0, "${"}
	tRight     = item{itemRightDelim, 0, "}"}
)

func tError(value string) item {
	return item{typ: itemError, val: value}
}

func tText(value string) item {
	return item{typ: itemText, val: value}
}

func tVariable(name string) item {
	return item{typ: itemVariable, val: name}
}

// equal compares items ignoring their position
func equal(i1, i2 []item) bool {
	if len(i1) != len(i2) {
		return false
	}
	for k := range i1 {
		if i1[k].typ != i2[k].typ {
			return false
		}
		if i1[k].val != i2[k].val {
			return false
		}
	}
	return true
}

func Test_item_String(t *testing.T) {
	type fields struct {
		typ itemType
		val string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"dash operator", fields{itemDash, "-"}, `OP: "-"`},
		{"error", fields{itemError, "error"}, `ERROR: "error"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := item{
				typ: tt.fields.typ,
				val: tt.fields.val,
			}
			if got := i.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
