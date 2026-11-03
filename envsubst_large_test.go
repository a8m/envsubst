package envsubst

import (
	"fmt"
	"strings"
	"testing"
)

// A large input with thousands of text and variable nodes must come out
// whole and in order.
func TestStringManyNodes(t *testing.T) {
	var input, expected strings.Builder
	for i := 0; i < 5000; i++ {
		fmt.Fprintf(&input, "line %d: $BAR ${BAR} ${ENVSUBST_TEST_UNSET:-x}\n", i)
		fmt.Fprintf(&expected, "line %d: bar bar x\n", i)
	}
	str, err := String(input.String())
	if err != nil {
		t.Fatal(err)
	}
	if str != expected.String() {
		t.Error("Expect a large input with many variables to be substituted in full")
	}
}
