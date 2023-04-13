package writer

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

const minReadBufferSize = 16

var bufsizes = []int{
	0, minReadBufferSize, 23, 32, 46, 64, 93, 128, 1024, 4096,
}

func TestWriters(t *testing.T) {
	var data [8192]byte

	for i := 0; i < len(data); i++ {
		data[i] = byte(' ' + i%('~'-' '))
	}
	w := new(bytes.Buffer)

	funcs := []func(w io.Writer, size int) BufferedWriter{
		NewDoubleBufWriterSize,
	}
	for _, f := range funcs {
		for i := 0; i < len(bufsizes); i++ {
			for j := 0; j < len(bufsizes); j++ {
				// i , j = 3,2
				nwrite := bufsizes[i]
				bs := bufsizes[j]

				// Write nwrite bytes using buffer size bs.
				// Check that the right amount makes it out
				// and that the data is correct.

				w.Reset()
				buf := f(w, bs)
				context := fmt.Sprintf("nwrite=%d bufsize=%d", nwrite, bs)
				n, e1 := buf.Write(data[0:nwrite])
				if e1 != nil || n != nwrite {
					t.Errorf("%s: buf.Write %d = %d, %v", context, nwrite, n, e1)
					continue
				}

				if e := buf.Flush(); e != nil {
					t.Errorf("%s: buf.Flush = %v", context, e)
				}

				written := w.Bytes()
				if len(written) != nwrite {
					t.Errorf("%s: %d bytes written", context, len(written))
				}
				for l := 0; l < len(written); l++ {
					if written[l] != data[l] {
						t.Errorf("wrong bytes written")
						t.Errorf("want=%q", data[0:len(written)])
						t.Errorf("have=%q", written)
					}
				}
			}
		}
	}
}
