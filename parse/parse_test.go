package parse

import (
	"os"
	"testing"
)

func init() {
	os.Setenv("BAR", "bar")
	os.Setenv("FOO", "foo")
	os.Setenv("EMPTY", "")
}

type parseTest struct {
	name     string
	input    string
	expected string
	hasErr   bool
}

var parseTests = []parseTest{
	{"empty", "", "", false},
	{"env only", "$BAR", "bar", false},
	{"with text", "$BAR baz", "bar baz", false},
	{"concatenated", "$BAR$FOO", "barfoo", false},
	{"2 env var", "$BAR - $FOO", "bar - foo", false},
	{"invalid var", "$_ bar", "$_ bar", false},
	{"invalid subst var", "${_} bar", "${_} bar", false},
	{"value of $var", "${BAR}baz", "barbaz", false},
	{"$var not set -", "${NOTSET-$BAR}", "bar", false},
	{"$var not set =", "${NOTSET=$BAR}", "bar", false},
	{"$var set but empty -", "${EMPTY-$BAR}", "", false},
	{"$var set but empty =", "${EMPTY=$BAR}", "", false},
	{"$var not set or empty :-", "${EMPTY:-$BAR}", "bar", false},
	{"$var not set or empty :=", "${EMPTY:=$BAR}", "bar", false},
	{"if $var set evaluate expression as $other +", "${EMPTY+hello}", "hello", false},
	{"if $var set evaluate expression as $other :+", "${EMPTY:+hello}", "hello", false},
	{"if $var not set, use empty string +", "${NOTSET+hello}", "", false},
	{"if $var not set, use empty string :+", "${NOTSET:+hello}", "", false},
	{"multi line string", "hello $BAR\nhello ${EMPTY:=$FOO}", "hello bar\nhello foo", false},
	// bad substitution
	{"closing brace expected", "hello ${", "", true},
}

func TestParse(t *testing.T) {
	for _, test := range parseTests {
		result, err := New(test.name).Parse(test.input)
		hasErr := err != nil
		if hasErr != test.hasErr {
			t.Errorf("%s=(error): got\n\t%v\nexpected\n\t%v", test.name, hasErr, test.hasErr)
		}
		if result != test.expected {
			t.Errorf("%s=(%q): got\n\t%v\nexpected\n\t%v", test.name, test.input, result, test.expected)
		}
	}
}
