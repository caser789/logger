package writer

import "time"

const (
	defaultBufSize     = 4096
	defaultFlushPeriod = 10 * time.Millisecond
)

type BufferedWriter interface {
	Write(p []byte) (n int, err error)
	Flush() error
}
