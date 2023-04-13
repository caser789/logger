package lumberjack

import (
	"io"

	"github.com/caser789/logger/internal/writer"
)

type WriterWrapper func(w io.Writer) writer.BufferedWriter

var (
	defaultWriterWrapper WriterWrapper
)

func init() {
	WithDoubleBufWrapper(4 * 1024)
}

func WithDoubleBufWrapper(size int) {
	defaultWriterWrapper = func(w io.Writer) writer.BufferedWriter {
		return writer.NewDoubleBufWriterSize(w, size)
	}
}
