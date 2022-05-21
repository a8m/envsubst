package envsubst

import (
	"io/ioutil"
	"os"

	"github.com/a8m/envsubst/parse"
)

// String returns the parsed template string after processing it.
// If the parser encounters invalid input, it returns an error describing the failure.
func String(s string) (string, error) {
	return stringRestricted(s, false, false, nil)
}

// StringSelectedEnvs returns the parsed template string for only the selected Envs
// specified by the caller after processing it.
// If the parser encounters invalid input, it returns an error describing the failure.
func StringSelectedEnvs(s string, selectedEnvs []string) (string, error) {
	return stringRestricted(s, false, false, selectedEnvs)
}

// StringRestricted returns the parsed template string after processing it.
// If the parser encounters invalid input, or a restriction is violated, it returns
// an error describing the failure.
// Errors on first failure or returns a collection of failures if failOnFirst is false
func StringRestricted(s string, noUnset, noEmpty bool) (string, error) {
	return stringRestricted(s, noUnset, noEmpty, nil)
}

// StringRestrictedSelectedEnvs returns the parsed template string for only the selected Envs
// specified by the caller after processing it.
// If the parser encounters invalid input, or a restriction is violated, it returns
// an error describing the failure.
// Errors on first failure or returns a collection of failures if failOnFirst is false
func StringRestrictedSelectedEnvs(s string, noUnset, noEmpty bool, selectedEnvs []string) (string, error) {
	return stringRestricted(s, noUnset, noEmpty, selectedEnvs)
}

func stringRestricted(s string, noUnset, noEmpty bool, selectedEnvs []string) (string, error) {
	return parse.New("string", os.Environ(),
		&parse.Restrictions{NoUnset: noUnset, NoEmpty: noEmpty}, selectedEnvs).Parse(s)
}

// Bytes returns the bytes represented by the parsed template after processing it.
// If the parser encounters invalid input, it returns an error describing the failure.
func Bytes(b []byte) ([]byte, error) {
	return bytesRestricted(b, false, false, nil)
}

// BytesSelectedEnvs returns the bytes represented by the parsed template for only the selected Envs
// specified by the caller after processing it.
// If the parser encounters invalid input, it returns an error describing the failure.
func BytesSelectedEnvs(b []byte, selectedEnvs []string) ([]byte, error) {
	return bytesRestricted(b, false, false, selectedEnvs)
}

// BytesRestricted returns the bytes represented by the parsed template after processing it.
// If the parser encounters invalid input, or a restriction is violated, it returns
// an error describing the failure.
func BytesRestricted(b []byte, noUnset, noEmpty bool) ([]byte, error) {
	return bytesRestricted(b, noUnset, noEmpty, nil)
}

// BytesRestrictedSelectedEnvs returns the bytes represented by the parsed template for
// only the selected Envs specified by the caller after processing it.
// If the parser encounters invalid input, or a restriction is violated, it returns
// an error describing the failure.
func BytesRestrictedSelectedEnvs(b []byte, noUnset, noEmpty bool, selectedEnvs []string) ([]byte, error) {
	return bytesRestricted(b, noUnset, noEmpty, selectedEnvs)
}

func bytesRestricted(b []byte, noUnset, noEmpty bool, selectedEnvs []string) ([]byte, error) {
	s, err := parse.New("bytes", os.Environ(),
		&parse.Restrictions{NoUnset: noUnset, NoEmpty: noEmpty}, selectedEnvs).Parse(string(b))
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

// ReadFile call io.ReadFile with the given file name.
// If the call to io.ReadFile failed it returns the error; otherwise it will
// call envsubst.Bytes with the returned content.
func ReadFile(filename string) ([]byte, error) {
	return readFileRestricted(filename, false, false, nil)
}

// ReadFileSelectedEnvs call io.ReadFile with the given file name.
// If the call to io.ReadFile failed it returns the error; otherwise it will
// call envsubst.Bytes with the returned content and selected Envs
// specified by the caller.
func ReadFileSelectedEnvs(filename string, selectedEnvs []string) ([]byte, error) {
	return readFileRestricted(filename, false, false, selectedEnvs)
}

// ReadFileRestricted calls io.ReadFile with the given file name.
// If the call to io.ReadFile failed it returns the error; otherwise it will
// call envsubst.Bytes with the returned content.
func ReadFileRestricted(filename string, noUnset, noEmpty bool) ([]byte, error) {
	return readFileRestricted(filename, noUnset, noUnset, nil)
}

// ReadFileRestrictedSelectedEnvs calls io.ReadFile with the given file name.
// If the call to io.ReadFile failed it returns the error; otherwise it will
// call envsubst.Bytes with the returned content and selected Envs
// specified by the caller.
func ReadFileRestrictedSelectedEnvs(filename string, noUnset, noEmpty bool, selectedEnvs []string) ([]byte, error) {
	return readFileRestricted(filename, noUnset, noUnset, selectedEnvs)
}

// ReadFileRestricted calls io.ReadFile with the given file name.
// If the call to io.ReadFile failed it returns the error; otherwise it will
// call envsubst.Bytes with the returned content and selected Envs
// if specified by the caller.
func readFileRestricted(filename string, noUnset, noEmpty bool, selectedEnvs []string) ([]byte, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return bytesRestricted(b, noUnset, noEmpty, selectedEnvs)
}
