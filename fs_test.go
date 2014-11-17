package fs

import (
	"io"
	"io/ioutil"
	"testing"

	"golang.org/x/net/context"
)

func TestOpen(t *testing.T) {
	f, err := ioutil.TempFile("", "fs-open-")
	if err != nil {
		t.Fatal(err)
	}
	name := f.Name()
	if _, err := io.WriteString(f, "hello world"); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	_, err = Open(context.Background(), name)
	if err != nil {
		t.Errorf("Open(%q): %v", name, err)
	}

	// TODO read and close
}
