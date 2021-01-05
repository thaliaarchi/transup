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
		c.Next()
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
		eq, err := c.Equal()
		if err != nil {
			return false, err
		}
		if !eq {
			if log != nil {
				src, dst := c.Sections()
				fmt.Fprintf(log, "%s section %d not equal\n", filename, i)
				d := hex.Dumper(log)
				io.Copy(d, src)
				d.Close()
				d = hex.Dumper(log)
				io.Copy(d, dst)
				d.Close()
			}
			return false, nil
		}
	}
	return true, nil
}

type Comparer struct {
	r      io.ReaderAt
	buf    bytes.Buffer
	off, n int64 // offset of current section in r
}

var _ io.ReadWriter = (*Comparer)(nil)

func NewComparer(r io.ReaderAt) *Comparer {
	return &Comparer{r: r}
}

func (c *Comparer) Read(b []byte) (n int, err error) {
	n, err = c.r.ReadAt(b, c.off+c.n)
	c.n += int64(n)
	return
}

func (c *Comparer) Write(b []byte) (n int, err error) {
	return c.buf.Write(b)
}

func (c *Comparer) Equal() (bool, error) {
	r := io.NewSectionReader(c.r, c.off, c.n)
	var rb [128]byte
	wb := c.buf.Bytes()
	for {
		n, err := r.Read(rb[:])
		if err == io.EOF {
			return len(wb) == 0, nil
		}
		if err != nil {
			return false, err
		}
		if len(wb) < n || !bytes.Equal(rb[:n], wb[:n]) {
			return false, nil
		}
		wb = wb[n:]
	}
}

func (c *Comparer) Sections() (src, dst io.Reader) {
	return io.NewSectionReader(c.r, c.off, c.n), &c.buf
}

func (c *Comparer) Next() {
	c.buf.Reset()
	c.off += c.n
	c.n = 0
}
