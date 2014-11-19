package fs

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"syscall"
	"testing"
	"time"

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

// TODO: test spurious signal does not cancel

// This test does not work on darwin. A write(2) that is partially started
// will be restarted even if the handler does not specify SA_RESTART. Ugh.
func TestWriteInterruptPipe(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	r, w, err := Pipe(ctx)
	if err != nil {
		t.Fatal(err)
	}
	w.SetNonBlocking()

	done := make(chan struct{}, 1)
	go func() {
		b := make([]byte, 1<<17)
		n, err := w.IO(ctx).Write(b)
		log.Printf("write done: %v, %v", n, err)
		done <- struct{}{}
	}()

	n, err := r.f.Read(make([]byte, 1024))
	log.Printf("read from pipe: %v, %v", n, err)

	log.Printf("calling cancel")
	cancel()
	log.Printf("canceled")
	time.Sleep(50 * time.Millisecond)
	log.Printf("slept, if not done yet, busted")
	//n, err = r.f.Read(make([]byte, 1<<20))
	//log.Printf("cheated, read out of pipa, %v, err=%v", n, err)
	<-done
	log.Printf("test done")

	w.f.Close()
	r.f.Close()
}

var signalCaught bool

func signalCaughtHandler(sig int32) {
	signalCaught = true
}

func TestInterruptRead(t *testing.T) {
	// Take over the interrupt handler, reset it when done.
	origHandler := intrHandler
	signalCaught = false
	intrHandler = signalCaughtHandler
	setsighandler()
	defer func() {
		intrHandler = origHandler
		setsighandler()
	}()

	// Cancel a raw pipe with a configured interrupt.
	ctx, cancel := context.WithCancel(context.Background())
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		var err error
		func() {
			defer interrupt(ctx)()
			_, err = r.Read(make([]byte, 1<<8))
		}()
		done <- err
	}()

	time.Sleep(50 * time.Millisecond) // TODO something better

	cancel()
	readErr := <-done

	if readErr.(*os.PathError).Err != syscall.EINTR {
		t.Errorf("not interrupted, got: %v", readErr)
	}

	if !signalCaught {
		t.Errorf("signal handler never called")
	}

	w.Close()
	r.Close()
}

func TestCancelRead(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Manually assemble the pipe to avoid setting it to non-blocking.
	osr, osw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	r, w := newFile(osr), newFile(osw)

	done := make(chan error, 1)
	go func() {
		_, err := r.IO(ctx).Read(make([]byte, 1<<8))
		done <- err
	}()

	time.Sleep(50 * time.Millisecond) // TODO something better

	cancel()
	readErr := <-done

	if readErr.(*os.PathError).Err != context.Canceled {
		t.Errorf("not canceled, got: %v", readErr)
	}

	w.f.Close()
	r.f.Close()
}
