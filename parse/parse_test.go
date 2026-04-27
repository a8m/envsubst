package parse

import (
	"testing"
)

var FakeEnv = []string{
	"BAR=bar",
	"FOO=foo",
	"EMPTY=",
	"ALSO_EMPTY=",
	"A=AAA",
}

type mode int

const (
	relaxed mode = iota
	noUnset
	noEmpty
	strict
)

var restrict = map[mode]*Restrictions{
	relaxed: Relaxed,
	noUnset: NoUnset,
	noEmpty: NoEmpty,
	strict:  Strict,
}

var errNone = map[mode]bool{}
var errUnset = map[mode]bool{noUnset: true, strict: true}
var errEmpty = map[mode]bool{noEmpty: true, strict: true}
var errAll = map[mode]bool{relaxed: true, noUnset: true, noEmpty: true, strict: true}
var errAllFull = map[mode]bool{relaxed: true, noUnset: true, noEmpty: true, strict: true}

type parseTest struct {
	name     string
	input    string
	expected string
	hasErr   map[mode]bool
}

var parseTests = []parseTest{
	{"empty", "", "", errNone},
	{"env only", "$BAR", "bar", errNone},
	{"with text", "$BAR baz", "bar baz", errNone},
	{"concatenated", "$BAR$FOO", "barfoo", errNone},
	{"2 env var", "$BAR - $FOO", "bar - foo", errNone},
	{"invalid var", "$_ bar", "$_ bar", errNone},
	{"invalid subst var", "${_} bar", "${_} bar", errNone},
	{"value of $var", "${BAR}baz", "barbaz", errNone},
	{"$var not set -", "${NOTSET-$BAR}", "bar", errNone},
	{"$var not set =", "${NOTSET=$BAR}", "bar", errNone},
	{"$var set but empty -", "${EMPTY-$BAR}", "", errEmpty},
	{"$var set but empty =", "${EMPTY=$BAR}", "", errEmpty},
	{"$var not set or empty :-", "${EMPTY:-$BAR}", "bar", errNone},
	{"$var not set or empty :=", "${EMPTY:=$BAR}", "bar", errNone},
	{"if $var set evaluate expression as $other +", "${EMPTY+hello}", "hello", errNone},
	{"if $var set evaluate expression as $other :+", "${EMPTY:+hello}", "hello", errNone},
	{"if $var not set, use empty string +", "${NOTSET+hello}", "", errNone},
	{"if $var not set, use empty string :+", "${NOTSET:+hello}", "", errNone},
	{"multi line string", "hello $BAR\nhello ${EMPTY:=$FOO}", "hello bar\nhello foo", errNone},
	{"issue #1", "${hello:=wo_rld} ${foo:=bar_baz}", "wo_rld bar_baz", errNone},
	{"issue #2", "name: ${NAME:=foo_qux}, key: ${EMPTY:=baz_bar}", "name: foo_qux, key: baz_bar", errNone},
	{"gh-issue-8", "prop=${HOME_URL-http://localhost:8080}", "prop=http://localhost:8080", errNone},
	// operators as leading values
	{"gh-issue-41-1", "${NOTSET--1}", "-1", errNone},
	{"gh-issue-41-2", "${NOTSET:--1}", "-1", errNone},
	{"gh-issue-41-3", "${NOTSET=-1}", "-1", errNone},
	{"gh-issue-41-4", "${NOTSET:==1}", "=1", errNone},

	// single letter
	{"gh-issue-43-1", "${A}", "AAA", errNone},

	// bad substitution
	{"closing brace expected", "hello ${", "", errAll},

	// test specifically for failure modes
	{"$var not set", "${NOTSET}", "", errUnset},
	{"$var set to empty", "${EMPTY}", "", errEmpty},
	// restrictions for plain variables without braces
	{"gh-issue-9", "$NOTSET", "", errUnset},
	{"gh-issue-9", "$EMPTY", "", errEmpty},

	{"$var and $DEFAULT not set -", "${NOTSET-$ALSO_NOTSET}", "", errUnset},
	{"$var and $DEFAULT not set :-", "${NOTSET:-$ALSO_NOTSET}", "", errUnset},
	{"$var and $DEFAULT not set =", "${NOTSET=$ALSO_NOTSET}", "", errUnset},
	{"$var and $DEFAULT not set :=", "${NOTSET:=$ALSO_NOTSET}", "", errUnset},
	{"$var and $OTHER not set +", "${NOTSET+$ALSO_NOTSET}", "", errNone},
	{"$var and $OTHER not set :+", "${NOTSET:+$ALSO_NOTSET}", "", errNone},

	{"$var empty and $DEFAULT not set -", "${EMPTY-$NOTSET}", "", errEmpty},
	{"$var empty and $DEFAULT not set :-", "${EMPTY:-$NOTSET}", "", errUnset},
	{"$var empty and $DEFAULT not set =", "${EMPTY=$NOTSET}", "", errEmpty},
	{"$var empty and $DEFAULT not set :=", "${EMPTY:=$NOTSET}", "", errUnset},
	{"$var empty and $OTHER not set +", "${EMPTY+$NOTSET}", "", errUnset},
	{"$var empty and $OTHER not set :+", "${EMPTY:+$NOTSET}", "", errUnset},

	{"$var not set and $DEFAULT empty -", "${NOTSET-$EMPTY}", "", errEmpty},
	{"$var not set and $DEFAULT empty :-", "${NOTSET:-$EMPTY}", "", errEmpty},
	{"$var not set and $DEFAULT empty =", "${NOTSET=$EMPTY}", "", errEmpty},
	{"$var not set and $DEFAULT empty :=", "${NOTSET:=$EMPTY}", "", errEmpty},
	{"$var not set and $OTHER empty +", "${NOTSET+$EMPTY}", "", errNone},
	{"$var not set and $OTHER empty :+", "${NOTSET:+$EMPTY}", "", errNone},

	{"$var and $DEFAULT empty -", "${EMPTY-$ALSO_EMPTY}", "", errEmpty},
	{"$var and $DEFAULT empty :-", "${EMPTY:-$ALSO_EMPTY}", "", errEmpty},
	{"$var and $DEFAULT empty =", "${EMPTY=$ALSO_EMPTY}", "", errEmpty},
	{"$var and $DEFAULT empty :=", "${EMPTY:=$ALSO_EMPTY}", "", errEmpty},
	{"$var and $OTHER empty +", "${EMPTY+$ALSO_EMPTY}", "", errEmpty},
	{"$var and $OTHER empty :+", "${EMPTY:+$ALSO_EMPTY}", "", errEmpty},

	// "$$" is preserved as literal text in this fork (see lex.go). Upstream
	// a8m/envsubst would collapse "$$" to "$" as a shell-style escape; we
	// don't, because it silently mangles inputs that contain literal "$$"
	// (e.g. KEDA CRD descriptions, Makefile-style snippets quoted in YAML).
	{"literal $$var", "FOO $$BAR BAZ", "FOO $$BAR BAZ", errNone},
	{"literal $${subst}", "FOO $${BAR} BAZ", "FOO $${BAR} BAZ", errNone},
	{"literal $$$var", "$$$BAR", "$$bar", errNone},
	{"literal $$${subst}", "$$${BAZ:-baz}", "$$baz", errNone},
	// "$$" inside a substitution operand is unaffected by this fork's lexer
	// change (lexSubstitution path is unchanged). Pinning the behavior so a
	// future refactor doesn't silently regress it.
	{"literal $$ in default operand", "${UNSET:-pre$$post}", "pre$$post", errNone},
	{"literal $$ in := default operand", "${UNSET:=pre$$post}", "pre$$post", errNone},

	// Combining literal $$ with a substitution — pins the documented forms in
	// README.md so docs and behavior can't drift apart silently.
	{"$$ then var triple-dollar", "$$$BAR", "$$bar", errNone},
	{"$$ then var space-braced", "$$ ${BAR}", "$$ bar", errNone},
	{"$$ then var no-braces does NOT substitute", "$$BAR", "$$BAR", errNone},
	{"$$ then var no-space-braced does NOT substitute", "$${BAR}", "$${BAR}", errNone},
}

var negativeParseTests = []parseTest{
	{"$NOTSET and EMPTY are displayed as in full error output", "${NOTSET} and $EMPTY", "variable ${NOTSET} not set\nvariable ${EMPTY} set but empty", errAllFull},
}

func TestParse(t *testing.T) {
	doTest(t, relaxed)
}

func TestParseNoUnset(t *testing.T) {
	doTest(t, noUnset)
}

func TestParseNoEmpty(t *testing.T) {
	doTest(t, noEmpty)
}

func TestParseStrict(t *testing.T) {
	doTest(t, strict)
}

func TestParseStrictNoFailFast(t *testing.T) {
	doNegativeAssertTest(t, strict)
}

func doTest(t *testing.T, m mode) {
	for _, test := range parseTests {
		result, err := New(test.name, FakeEnv, restrict[m]).Parse(test.input)
		hasErr := err != nil
		if hasErr != test.hasErr[m] {
			t.Errorf("%s=(error): got\n\t%v\nexpected\n\t%v\ninput: %s\nresult: %s\nerror: %v",
				test.name, hasErr, test.hasErr[m], test.input, result, err)
		}
		if result != test.expected {
			t.Errorf("%s=(%q): got\n\t%v\nexpected\n\t%v", test.name, test.input, result, test.expected)
		}
	}
}

func doNegativeAssertTest(t *testing.T, m mode) {
	for _, test := range negativeParseTests {
		result, err := (*&Parser{Name: test.name, Env: FakeEnv, Restrict: restrict[m], Mode: AllErrors}).Parse(test.input)
		hasErr := err != nil
		if hasErr != test.hasErr[m] {
			t.Errorf("%s=(error): got\n\t%v\nexpected\n\t%v\ninput: %s\nresult: %s\nerror: %v",
				test.name, hasErr, test.hasErr[m], test.input, result, err)
		}
		if err.Error() != test.expected {
			t.Errorf("%s=(%q): got\n\t%v\nexpected\n\t%v", test.name, test.input, err.Error(), test.expected)
		}
	}
}

// Test VarFilter prefix filtering
func TestVarFilter(t *testing.T) {
	tests := []struct {
		name     string
		filter   *VarFilter
		varName  string
		expected bool
	}{
		{"nil filter allows all", nil, "ANY_VAR", true},
		{"nil filter allows ARGOCD_ENV_", nil, "ARGOCD_ENV_FOO", true},
		{"prefix match", &VarFilter{Prefixes: []string{"ARGOCD_ENV_"}}, "ARGOCD_ENV_FOO", true},
		{"prefix match nested", &VarFilter{Prefixes: []string{"ARGOCD_ENV_"}}, "ARGOCD_ENV_CLUSTER_NAME", true},
		{"prefix no match", &VarFilter{Prefixes: []string{"ARGOCD_ENV_"}}, "OTHER_VAR", false},
		{"prefix no match partial", &VarFilter{Prefixes: []string{"ARGOCD_ENV_"}}, "ARGOCD_ENV", false},
		{"multiple prefixes match first", &VarFilter{Prefixes: []string{"FOO_", "BAR_"}}, "FOO_VAR", true},
		{"multiple prefixes match second", &VarFilter{Prefixes: []string{"FOO_", "BAR_"}}, "BAR_VAR", true},
		{"multiple prefixes no match", &VarFilter{Prefixes: []string{"FOO_", "BAR_"}}, "BAZ_VAR", false},
		{"empty prefixes blocks all", &VarFilter{Prefixes: []string{}}, "ANY_VAR", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.filter.IsAllowed(test.varName)
			if result != test.expected {
				t.Errorf("VarFilter.IsAllowed(%q) = %v, expected %v", test.varName, result, test.expected)
			}
		})
	}
}

// Test parsing with prefix filter
func TestParseWithPrefixFilter(t *testing.T) {
	env := []string{
		"ARGOCD_ENV_FOO=foo",
		"ARGOCD_ENV_BAR=bar",
		"OTHER_VAR=other",
	}

	filter := &VarFilter{Prefixes: []string{"ARGOCD_ENV_"}}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic prefix filtering
		{"allowed var substituted", "$ARGOCD_ENV_FOO", "foo"},
		{"disallowed var kept literal", "$OTHER_VAR", "$OTHER_VAR"},
		{"mixed vars", "$ARGOCD_ENV_FOO $OTHER_VAR", "foo $OTHER_VAR"},

		// Braced syntax
		{"braced allowed var", "${ARGOCD_ENV_FOO}", "foo"},
		{"braced disallowed var", "${OTHER_VAR}", "${OTHER_VAR}"},

		// Default values - KEY TEST CASE
		{"unset allowed var with default", "${ARGOCD_ENV_UNSET:-fallback}", "fallback"},
		{"unset disallowed var with default", "${OTHER_UNSET:-fallback}", "${OTHER_UNSET:-fallback}"},
		{"set allowed var with default", "${ARGOCD_ENV_FOO:-fallback}", "foo"},

		// Complex defaults
		{"allowed var empty default", "${ARGOCD_ENV_BAR:-}", "bar"},
		{"allowed var text default", "${ARGOCD_ENV_MISSING:-default_value}", "default_value"},

		// Multiple vars
		{"all allowed", "$ARGOCD_ENV_FOO ${ARGOCD_ENV_BAR}", "foo bar"},
		{"all disallowed", "$OTHER_VAR ${ANOTHER}", "$OTHER_VAR ${ANOTHER}"},
		{"mixed with defaults", "${ARGOCD_ENV_UNSET:-def1} ${OTHER_UNSET:-def2}", "def1 ${OTHER_UNSET:-def2}"},

		// Text around vars
		{"text with allowed", "hello $ARGOCD_ENV_FOO world", "hello foo world"},
		{"text with disallowed", "hello $OTHER_VAR world", "hello $OTHER_VAR world"},

		// Real world ArgoCD pattern
		{"argocd health check pattern", "path: ${ARGOCD_ENV_HEALTH_PATH:-/ping}", "path: /ping"},
		{"argocd service name pattern", "name: ${ARGOCD_ENV_SERVICE_NAME:-$ARGOCD_ENV_FOO}", "name: foo"},

		// KEDA CRD scenario: upstream chart embeds Kubernetes' own env-var
		// expansion docs in CRD schema descriptions. Those strings contain
		// literal "$$" that must NOT be collapsed, or ArgoCD shows perpetual
		// drift between desired ($) and live ($$). See parse/lex.go.
		{"keda crd description literal $$", "description: Double $$ are reduced to a single $", "description: Double $$ are reduced to a single $"},
		{"keda kubelet escape literal", "value: $$(VAR_NAME) literal", "value: $$(VAR_NAME) literal"},
		{"literal $$ alongside prefix var", "$ARGOCD_ENV_FOO sees Double $$ are reduced", "foo sees Double $$ are reduced"},
		{"literal $$ alongside default", "${ARGOCD_ENV_HEALTH_PATH:-/ping} and $$VAR", "/ping and $$VAR"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := &Parser{
				Name:      test.name,
				Env:       env,
				Restrict:  Relaxed,
				VarFilter: filter,
			}
			result, err := p.Parse(test.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != test.expected {
				t.Errorf("Parse(%q) = %q, expected %q", test.input, result, test.expected)
			}
		})
	}
}

// Test all substitution operators with prefix filter
func TestPrefixFilterAllOperators(t *testing.T) {
	env := []string{
		"ARGOCD_ENV_SET=set_value",
		"ARGOCD_ENV_EMPTY=",
		"OTHER_SET=other_value",
		"OTHER_EMPTY=",
	}

	filter := &VarFilter{Prefixes: []string{"ARGOCD_ENV_"}}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// === ALLOWED PREFIX - Simple $VAR ===
		{"allowed simple set", "$ARGOCD_ENV_SET", "set_value"},
		{"allowed simple unset", "$ARGOCD_ENV_UNSET", ""},
		{"allowed simple empty", "$ARGOCD_ENV_EMPTY", ""},

		// === ALLOWED PREFIX - Braced ${VAR} ===
		{"allowed braced set", "${ARGOCD_ENV_SET}", "set_value"},
		{"allowed braced unset", "${ARGOCD_ENV_UNSET}", ""},
		{"allowed braced empty", "${ARGOCD_ENV_EMPTY}", ""},

		// === ALLOWED PREFIX - Default if unset or empty :- ===
		{"allowed :- set", "${ARGOCD_ENV_SET:-default}", "set_value"},
		{"allowed :- unset", "${ARGOCD_ENV_UNSET:-default}", "default"},
		{"allowed :- empty", "${ARGOCD_ENV_EMPTY:-default}", "default"},

		// === ALLOWED PREFIX - Default if unset only - ===
		{"allowed - set", "${ARGOCD_ENV_SET-default}", "set_value"},
		{"allowed - unset", "${ARGOCD_ENV_UNSET-default}", "default"},
		{"allowed - empty", "${ARGOCD_ENV_EMPTY-default}", ""}, // empty stays empty

		// === ALLOWED PREFIX - Assign default := ===
		{"allowed := set", "${ARGOCD_ENV_SET:=default}", "set_value"},
		{"allowed := unset", "${ARGOCD_ENV_UNSET:=default}", "default"},
		{"allowed := empty", "${ARGOCD_ENV_EMPTY:=default}", "default"},

		// === ALLOWED PREFIX - Alternate if set :+ ===
		{"allowed :+ set", "${ARGOCD_ENV_SET:+alternate}", "alternate"},
		{"allowed :+ unset", "${ARGOCD_ENV_UNSET:+alternate}", ""},

		// === ALLOWED PREFIX - Variable as default (level 1 nesting) ===
		{"allowed var default same prefix", "${ARGOCD_ENV_UNSET:-$ARGOCD_ENV_SET}", "set_value"},
		{"allowed var default diff prefix", "${ARGOCD_ENV_UNSET:-$OTHER_SET}", "$OTHER_SET"}, // inner not substituted

		// === DISALLOWED PREFIX - Should stay literal ===
		{"disallowed simple", "$OTHER_SET", "$OTHER_SET"},
		{"disallowed braced", "${OTHER_SET}", "${OTHER_SET}"},
		{"disallowed with default", "${OTHER_UNSET:-default}", "${OTHER_UNSET:-default}"},
		{"disallowed var default", "${OTHER_UNSET:-$OTHER_SET}", "${OTHER_UNSET:-$OTHER_SET}"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := &Parser{
				Name:      test.name,
				Env:       env,
				Restrict:  Relaxed,
				VarFilter: filter,
			}
			result, err := p.Parse(test.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != test.expected {
				t.Errorf("Parse(%q) = %q, expected %q", test.input, result, test.expected)
			}
		})
	}
}

// TestNestedDefaultsLimitation documents that deeply nested ${VAR:-${VAR2:-default}}
// is a known limitation of the parser. Level 1 nesting works, deeper levels have issues.
func TestNestedDefaultsLimitation(t *testing.T) {
	env := []string{"A=a", "B=b"}

	// Level 1 nesting works
	t.Run("level 1 nesting works", func(t *testing.T) {
		p := &Parser{Name: "test", Env: env, Restrict: Relaxed}
		result, _ := p.Parse("${UNSET:-$A}")
		if result != "a" {
			t.Errorf("Level 1 nesting failed: got %q, expected %q", result, "a")
		}
	})

	// Level 2+ has known issues with extra closing braces
	// This is a pre-existing limitation in a8m/envsubst
	t.Run("level 2 nesting limitation", func(t *testing.T) {
		t.Skip("Known limitation: nested ${VAR:-${VAR2:-default}} produces extra }")
		p := &Parser{Name: "test", Env: env, Restrict: Relaxed}
		result, _ := p.Parse("${UNSET1:-${UNSET2:-$B}}")
		if result != "b" {
			t.Errorf("Level 2 nesting: got %q, expected %q", result, "b")
		}
	})
}

// Test that nil filter allows all vars (backward compatibility)
func TestParseWithNilFilter(t *testing.T) {
	env := []string{
		"FOO=foo",
		"BAR=bar",
	}

	p := &Parser{
		Name:      "nil-filter",
		Env:       env,
		Restrict:  Relaxed,
		VarFilter: nil, // nil filter = allow all
	}

	input := "$FOO ${BAR} ${UNSET:-default}"
	expected := "foo bar default"

	result, err := p.Parse(input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != expected {
		t.Errorf("Parse(%q) = %q, expected %q", input, result, expected)
	}
}
