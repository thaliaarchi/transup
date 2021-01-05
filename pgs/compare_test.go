package pgs

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestTwoWay(t *testing.T) {
	files, err := filepath.Glob("../testdata/**/*.sup")
	if err != nil {
		t.Fatal(err)
	}
	for _, filename := range files {
		eq, err := testFileTwoWay(filename, os.Stderr)
		if err != nil {
			t.Error(err)
		} else if !eq {
			t.Errorf("%s differs", filename)
			break
		}
	}
}

func testFileTwoWay(filename string, log io.Writer) (bool, error) {
	f, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer f.Close()
	c := NewComparer(f)
	r := NewReader(c)
	w := NewWriter(c)

	for i := 0; ; i++ {
		p, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, err
		}
		if err := w.Write(p); err != nil {
			return false, err
		}
		if !c.Equal() {
			if log != nil {
				rbuf, wbuf := c.Buffers()
				fmt.Fprintf(log, "%s section %d not equal\n", filename, i)
				fmt.Fprintln(log, "raw:")
				d := hex.Dumper(log)
				io.Copy(d, rbuf)
				d.Close()
				fmt.Fprintln(log, "serialized:")
				d = hex.Dumper(log)
				io.Copy(d, wbuf)
				d.Close()
			}
			return false, nil
		}
		c.Reset()
	}
	return true, nil
}

type Comparer struct {
	r          io.Reader
	rbuf, wbuf bytes.Buffer
}

var _ io.ReadWriter = (*Comparer)(nil)

func NewComparer(r io.Reader) *Comparer {
	var c Comparer
	c.r = io.TeeReader(r, &c.rbuf)
	return &c
}

func (c *Comparer) Read(b []byte) (n int, err error) {
	return c.r.Read(b)
}

func (c *Comparer) Write(b []byte) (n int, err error) {
	return c.wbuf.Write(b)
}

func (c *Comparer) Equal() bool {
	return bytes.Equal(c.rbuf.Bytes(), c.wbuf.Bytes())
}

func (c *Comparer) Buffers() (rbuf, wbuf io.Reader) {
	return &c.rbuf, &c.wbuf
}

func (c *Comparer) Reset() {
	c.rbuf.Reset()
	c.wbuf.Reset()
}
