package fs

import (
	"bytes"
	"io/ioutil"
	"testing"

	"golang.org/x/net/context"
)

func writeTempFile(t *testing.T, data []byte) (fileName string) {
	f, err := ioutil.TempFile("", "fs-open-")
	if err != nil {
		t.Fatal(err)
	}
	name := f.Name()
	if _, err := f.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return name
}

func TestOpen(t *testing.T) {
	want := []byte("hello world")
	path := writeTempFile(t, want)

	f, err := Open(context.Background(), path)
	if err != nil {
		t.Errorf("Open(%q): %v", path, err)
	}
	got, err := ioutil.ReadAll(f.IO(context.Background()))
	if err != nil {
		t.Errorf("read with background context: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("got %q, wrote %q", string(got), string(want))
	}
}

/*
func TestReadInterrupt(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
}
*/
