package ar

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"strconv"
	"strings"
	"time"
)

const (
	headerSize       = 60
	PrefixSymdef     = "__.SYMDEF"
	bsdVariantMarker = "#1/"
)

var (
	MagicHeader      = []byte("!<arch>\n")
	ErrInvalidFormat = errors.New("not ar file format")
)

type File struct {
	*io.SectionReader
	Header
}

// https://en.wikipedia.org/wiki/Ar_(Unix)
type Header struct {
	Name     string
	Size     int64
	ModTime  time.Time
	UID      int
	GID      int
	Mode     fs.FileMode
	nameSize int64
}

type Iter struct {
	sr *io.SectionReader
}

func NewArchive(ra io.ReaderAt) ([]*File, error) {
	iter, err := NewIter(ra)
	if err != nil {
		return nil, err
	}

	files := []*File{}
	for file, err := range iter.Next() {
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func NewIter(r io.ReaderAt) (*Iter, error) {
	buf := make([]byte, len(MagicHeader))
	sr := io.NewSectionReader(r, 0, 1<<63-1)
	if _, err := io.ReadFull(sr, buf); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, ErrInvalidFormat
		}
		return nil, err
	}

	if !bytes.Equal(MagicHeader, buf) {
		return nil, fmt.Errorf("invalid magic header want: %s, got: %s: %w",
			string(MagicHeader), string(buf), ErrInvalidFormat)
	}

	return &Iter{sr: sr}, nil
}

func (r *Iter) Next() iter.Seq2[*File, error] {
	return func(yield func(*File, error) bool) {
		cur := int64(len(MagicHeader))
		for {
			f, err := load(r.sr, cur)
			if errors.Is(err, io.EOF) {
				return
			}

			if !yield(f, err) {
				return
			}
			if err != nil {
				return
			}
			cur += f.Header.Size + headerSize
		}
	}
}

func load(sr *io.SectionReader, off int64) (*File, error) {
	hdr, err := readHeader(sr, off)
	if err != nil {
		return nil, err
	}

	filesr := io.NewSectionReader(sr,
		off+headerSize+hdr.nameSize,
		hdr.Size-hdr.nameSize)
	f := &File{SectionReader: filesr, Header: *hdr}
	return f, nil
}

func readHeader(sr *io.SectionReader, off int64) (*Header, error) {
	var hdrBuf [headerSize]byte
	hdrsr := io.NewSectionReader(sr, off, headerSize)
	n, err := io.ReadFull(hdrsr, hdrBuf[:])
	if err != nil {
		return nil, err
	}

	if n != headerSize {
		return nil, fmt.Errorf("error reading header want: %d bytes, got: %d bytes", headerSize, n)
	}

	hdr, err := parseHeader(hdrBuf)
	if err != nil {
		return nil, err
	}

	// handle BSD variant
	if strings.HasPrefix(hdr.Name, bsdVariantMarker) {
		trimmedSize := strings.TrimPrefix(hdr.Name, bsdVariantMarker)
		parsedSize, err := parseDecimal(trimmedSize)
		if err != nil {
			return nil, err
		}

		namesr := io.NewSectionReader(sr, off+headerSize, parsedSize)
		nameBuf := make([]byte, parsedSize)
		if _, err := io.ReadFull(namesr, nameBuf); err != nil {
			return nil, err
		}

		// update
		hdr.Name = strings.TrimRight(string(nameBuf), "\x00")
		hdr.nameSize = int64(parsedSize)
	}

	return hdr, nil
}

func parseHeader(buf [headerSize]byte) (*Header, error) {
	name := TrimTailSpace(buf[0:16])

	parsedMTime, err := parseDecimal(TrimTailSpace(buf[16:28]))
	if err != nil {
		return nil, fmt.Errorf("parse modtime: %w", err)
	}
	modTime := time.Unix(parsedMTime, 0)

	parsedUID, err := parseDecimal(TrimTailSpace(buf[28:34]))
	if err != nil {
		return nil, fmt.Errorf("parse uid: %w", err)
	}

	parsedGID, err := parseDecimal(TrimTailSpace(buf[34:40]))
	if err != nil {
		return nil, fmt.Errorf("parse gid: %w", err)
	}

	uid, gid := int(parsedUID), int(parsedGID)

	parsedPerm, err := parseOctal(TrimTailSpace(buf[40:48]))
	if err != nil {
		return nil, fmt.Errorf("parse mode: %w", err)
	}

	perm := fs.FileMode(parsedPerm)

	size, err := parseDecimal(TrimTailSpace(buf[48:58]))
	if err != nil {
		return nil, fmt.Errorf("parse size value of name: %w", err)
	}

	endChars := buf[58:60]
	if want := []byte{0x60, 0x0a}; !bytes.Equal(want, endChars) {
		return nil, fmt.Errorf("unexpected ending characters want: %x, got: %x", want, endChars)
	}

	return &Header{
		Size:    size,
		Name:    name,
		ModTime: modTime,
		GID:     gid,
		UID:     uid,
		Mode:    perm,
	}, nil
}

func parseDecimal(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func parseOctal(s string) (int64, error) {
	return strconv.ParseInt(s, 8, 64)
}

func TrimTailSpace(b []byte) string {
	return strings.TrimRight(string(b), " ")
}
