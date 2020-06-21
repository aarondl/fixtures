package fixtures

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"unicode"

	"github.com/pmezard/go-difflib/difflib"
)

const (
	dirPerms  = 0o775
	filePerms = 0o664

	fixtureDir    = "testdata"
	fixturePrefix = "fixture."
)

var (
	flagUpdateFixtures bool
)

func init() {
	flag.BoolVar(&flagUpdateFixtures, "fix", false, "Update fixtures")
}

// JSON takes an object as input and jsonifies it to create a fixture.
func JSON(t *testing.T, filename string, now interface{}) {
	t.Helper()

	var gotJSON []byte
	var err error
	gotJSON, err = json.MarshalIndent(now, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	bytesHelper(t, filename, gotJSON, false)
}

// String is a helper to avoid bytes.
func String(t *testing.T, filename string, now string) {
	t.Helper()
	bytesHelper(t, filename, []byte(now), false)
}

// Bytes compares the file to bytes
func Bytes(t *testing.T, filename string, now []byte) {
	t.Helper()
	bytesHelper(t, filename, now, true)
}

func bytesHelper(t *testing.T, filename string, now []byte, tryJSONFormat bool) {
	t.Helper()
	filename = filepath.Join(fixtureDir, fixturePrefix+filename)

	var err error
	if flagUpdateFixtures {
		if _, err = os.Stat(fixtureDir); os.IsNotExist(err) {
			err = os.Mkdir(fixtureDir, dirPerms)
			if err != nil {
				panic(err)
			}
		}
		err = ioutil.WriteFile(filename, now, filePerms)
		if err != nil {
			panic(err)
		}
		return
	}

	old, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatalf("fixture file does not exist: %s", filename)
	}

	nowIsBinary := false
	for _, c := range string(now) {
		if !unicode.IsPrint(c) && !unicode.IsSpace(c) {
			nowIsBinary = true
			break
		}
	}

	oldIsBinary := false
	for _, c := range string(old) {
		if !unicode.IsPrint(c) && !unicode.IsSpace(c) {
			oldIsBinary = true
			break
		}
	}

	if nowIsBinary && oldIsBinary {
		if !bytes.Equal(old, now) {
			t.Errorf("wrong value:\nwant:\n%x\ngot:\n%x\n", old, now)
		}
		return
	}

	// Attempt to format the ascii strings as json to enhance diff-ability
	// It is a very common case that String() or Bytes() are called.
	if tryJSONFormat {
		oldOut := new(bytes.Buffer)
		if err := json.Indent(oldOut, old, "", "  "); err == nil {
			nowOut := new(bytes.Buffer)
			if err := json.Indent(nowOut, now, "", "  "); err == nil {
				old = oldOut.Bytes()
				now = nowOut.Bytes()
			}
		}
	}

	udiff := difflib.UnifiedDiff{
		Context:  3,
		A:        difflib.SplitLines(string(old)),
		B:        difflib.SplitLines(string(now)),
		FromFile: filename,
		ToFile:   filename,
	}
	udiffStr, err := difflib.GetUnifiedDiffString(udiff)
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(old, now) {
		t.Errorf("wrong value:\n%s\n", udiffStr)
	}
}
