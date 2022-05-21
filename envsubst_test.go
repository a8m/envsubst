package envsubst

import (
	"io/ioutil"
	"os"
	"testing"
)

func init() {
	os.Setenv("BAR", "bar")
}

// Basic integration tests. because we  already test the
// templating processing in envsubst/parse;
func TestIntegration(t *testing.T) {
	input, expected := "foo $BAR", "foo bar"
	str, err := String(input)
	if err != nil {
		t.Error("String ran unsuccesfully")
	}
	if str != expected || err != nil {
		t.Error("Expect String integration test to pass")
	}
	strenv, err := StringSelectedEnvs(input, []string{"BAR"})
	if err != nil {
		t.Error("StringSelectedEnvs ran unsuccesfully")
	}
	if strenv != expected || err != nil {
		t.Error("Expect StringSelectedEnvs integration test to pass")
	}
	bytes, err := Bytes([]byte(input))
	if err != nil {
		t.Error("Bytes ran unsuccesfully")
	}
	if string(bytes) != expected || err != nil {
		t.Error("Expect bytes integration test to pass")
	}
	bytesenv, err := BytesSelectedEnvs([]byte(input), []string{"BAR"})
	if err != nil {
		t.Error("BytesSelectedEnvs ran unsuccesfully")
	}
	if string(bytesenv) != expected || err != nil {
		t.Error("Expect BytesSelectedEnvs integration test to pass")
	}
	readfile, err := ReadFile("testdata/file.tmpl")
	if err != nil {
		t.Error("ReadFile ran unsuccesfully")
	}
	fexpected, err := ioutil.ReadFile("testdata/file.out")
	if string(readfile) != string(fexpected) || err != nil {
		t.Error("Expect ReadFile integration test to pass")
	}
	readfileenvs, err := ReadFileSelectedEnvs("testdata/file-env.tmpl", []string{"BAR", "FOO", "BAZ", "ENV"})
	if err != nil {
		t.Error("ReadFileSelectedEnvs ran unsuccesfully")
	}
	fenvexpected, err := ioutil.ReadFile("testdata/file-env.out")
	if string(readfileenvs) != string(fenvexpected) || err != nil {
		t.Error("Expect ReadFileSelectedEnvs integration test to pass")
	}
}
