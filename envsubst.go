package envsubst

import (
	"io/ioutil"
	"os"

	"github.com/a8m/envsubst/parse"
)

// String returns the parsed template string after processing it.
// If the parser encounters invalid input, it returns an error describing the failure.
func String(s string) (string, error) {
	return StringRestricted(s, false, false)
}

// StringRestricted returns the parsed template string after processing it.
// If the parser encounters invalid input, or a restriction is violated, it returns
// an error describing the failure.
func StringRestricted(s string, noUnset, noEmpty bool) (string, error) {
	return parse.New("string", os.Environ(),
		&parse.Restrictions{noUnset, noEmpty}).Parse(s)
}

// Bytes returns the bytes represented by the parsed template after processing it.
// If the parser encounters invalid input, it returns an error describing the failure.
func Bytes(b []byte) ([]byte, error) {
	return BytesRestricted(b, false, false)
}

// BytesRestricted returns the bytes represented by the parsed template after processing it.
// If the parser encounters invalid input, or a restriction is violated, it returns
// an error describing the failure.
func BytesRestricted(b []byte, noUnset, noEmpty bool) ([]byte, error) {
	s, err := parse.New("bytes", os.Environ(),
		&parse.Restrictions{noUnset, noEmpty}).Parse(string(b))
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

// ReadFile call io.ReadFile with the given file name.
// If the call to io.ReadFile failed it returns the error; otherwise it will
// call envsubst.Bytes with the returned content.
func ReadFile(filename string) ([]byte, error) {
	return ReadFileRestricted(filename, false, false)
}

// ReadFileRestricted calls io.ReadFile with the given file name.
// If the call to io.ReadFile failed it returns the error; otherwise it will
// call envsubst.Bytes with the returned content.
func ReadFileRestricted(filename string, noUnset, noEmpty bool) ([]byte, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return BytesRestricted(b, noUnset, noEmpty)
}
