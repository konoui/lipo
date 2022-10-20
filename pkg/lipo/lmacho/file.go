package lmacho

import (
	"io"
)

type File struct {
	sr *io.SectionReader
	c  io.Closer
}

func (f *File) Read(p []byte) (int, error) {
	return f.sr.Read(p)
}

func (f *File) ReadAt(p []byte, off int64) (int, error) {
	return f.sr.ReadAt(p, off)
}

func (f *File) Close() error {
	return f.c.Close()
}
