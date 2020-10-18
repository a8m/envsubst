package parse

import (
	"os"
	"strings"
)

type Env interface {
	// Get fetches the value associated
	// with the key or an empty string
	// if nothing is found.
	Get(name string) string
	// Has checks if the given key is present
	// in the datasource.
	Has(name string) bool
	// Lookup searches the datasource(s)
	// for a possible match. Returns the
	// value associated with the key and
	// the status of the search.
	Lookup(name string) (string, bool)
}

type FunctorEnv struct {
	Mapping func(string) string
}

func NewFunctorEnv(mapping func(string) string) *FunctorEnv {
	return &FunctorEnv {
		Mapping: mapping,
	}
}

// NewOsFunctorEnv will create a FunctorEnv
// object using os.Getenv as datasource.
func NewOsFunctorEnv() *FunctorEnv {
	return &FunctorEnv {
		Mapping: os.Getenv,
	}
}

// Get is a wrapper around Lookup,
// only returning the first return value.
func (e *FunctorEnv) Get(name string) string {
	v, _ := e.Lookup(name)
	return v
}

// Has is a wrapper around Lookup,
// only returning the second return value.
func (e *FunctorEnv) Has(name string) bool {
	_, ok := e.Lookup(name)
	return ok
}

// Lookup returns the value associated with the provided key.
// If the lookup yields an empty string, the behaviour
// is the same as if the key does not exist in the first place:
// the search fails.
func (e *FunctorEnv) Lookup(name string) (string, bool) {
	if value := e.Mapping(name); value != "" {
		return value, true
	}

	return "", false
}

type SliceEnv []string

func NewSliceEnv(source []string) SliceEnv {
	return SliceEnv(source)
}

// NewOsSliceEnv will create a SliceEnv
// object using os.Environ() as datasource.
func NewOsSliceEnv() SliceEnv {
	return SliceEnv(os.Environ())
}

// Get is a wrapper around Lookup,
// only returning the first return value.
func (e SliceEnv) Get(name string) string {
	v, _ := e.Lookup(name)
	return v
}

// Has is a wrapper around Lookup,
// only returning the second return value.
func (e SliceEnv) Has(name string) bool {
	_, ok := e.Lookup(name)
	return ok
}

// Lookup will iterate over the date source,
// looking for elements with the given key
// and an equal sign (=) as prefix.
func (e SliceEnv) Lookup(name string) (string, bool) {
	prefix := name + "="
	for _, pair := range e {
		if strings.HasPrefix(pair, prefix) {
			return pair[len(prefix):], true
		}
	}
	return "", false
}
