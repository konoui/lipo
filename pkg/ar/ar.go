package ar

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strconv"
	"strings"
	"time"
)

const (
	headerSize   = 60
	PrefixSymdef = "__.SYMDEF"
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
// TODO other fields
type Header struct {
	Name     string
	Size     int64
	ModTime  time.Time
	UID      int
	GID      int
	Mode     fs.FileMode
	nameSize int64
}

type Reader struct {
	sr   *io.SectionReader
	cur  int64
	next int64
}

// NewArchive is a wrapper
func NewArchive(ra io.ReaderAt) ([]*File, error) {
	r, err := NewReader(ra)
	if err != nil {
		return nil, err
	}

	files := []*File{}

	for {
		f, err := r.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		files = append(files, f)
	}
	return files, nil
}

func NewReader(r io.ReaderAt) (*Reader, error) {
	mhLen := len(MagicHeader)
	buf := make([]byte, mhLen)
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

	return &Reader{sr: sr, cur: int64(mhLen), next: int64(mhLen)}, nil
}

// Next returns a file header and a reader of original data
func (r *Reader) Next() (*File, error) {
	hdr, err := r.readHeader()
	if err != nil {
		return nil, err
	}

	sr := io.NewSectionReader(r.sr, r.cur+hdr.nameSize, hdr.Size-hdr.nameSize)

	r.cur += hdr.Size
	return &File{
		SectionReader: sr,
		Header:        *hdr,
	}, nil
}

func (r *Reader) readHeader() (*Header, error) {
	if _, err := r.sr.Seek(r.next, io.SeekStart); err != nil {
		return nil, err
	}

	header := make([]byte, headerSize)
	n, err := io.ReadFull(r.sr, header)
	if err != nil {
		return nil, err
	}

	if n != headerSize {
		return nil, fmt.Errorf("error reading header want: %d bytes, got: %d bytes", headerSize, n)
	}

	name := TrimTailSpace(header[0:16])

	parsedMTime, err := parseDecimal(TrimTailSpace(header[16:28]))
	if err != nil {
		return nil, fmt.Errorf("parse mtime: %w", err)
	}
	modTime := time.Unix(parsedMTime, 0)

	parsedUID, err := parseDecimal(TrimTailSpace(header[28:34]))
	if err != nil {
		return nil, fmt.Errorf("parse uid: %w", err)
	}

	parsedGID, err := parseDecimal(TrimTailSpace(header[34:40]))
	if err != nil {
		return nil, fmt.Errorf("parse gid: %w", err)
	}

	uid, gid := int(parsedUID), int(parsedGID)

	parsedPerm, err := parseOctal(TrimTailSpace(header[40:48]))
	if err != nil {
		return nil, fmt.Errorf("parse mode: %w", err)
	}

	perm := fs.FileMode(parsedPerm)

	size, err := parseDecimal(TrimTailSpace(header[48:58]))
	if err != nil {
		return nil, fmt.Errorf("parse size value of name: %w", err)
	}

	endChars := header[58:60]
	if want := []byte{0x60, 0x0a}; !bytes.Equal(want, endChars) {
		return nil, fmt.Errorf("unexpected ending characters want: %x, got: %x", want, endChars)
	}

	// update
	r.cur += headerSize

	var nameSize int64 = 0
	// handle BSD variant
	if strings.HasPrefix(name, "#1/") {
		trimmedSize := strings.TrimPrefix(name, "#1/")
		parsedSize, err := parseDecimal(trimmedSize)
		if err != nil {
			return nil, err
		}

		nameBuf := make([]byte, parsedSize)
		if _, err := io.ReadFull(r.sr, nameBuf); err != nil {
			return nil, err
		}

		// update
		name = strings.TrimRight(string(nameBuf), "\x00")
		// update
		nameSize = int64(parsedSize)
	}

	// align to read body
	if size%2 != 0 {
		if _, err := io.CopyN(io.Discard, r.sr, 1); err != nil {
			if !errors.Is(err, io.EOF) {
				return nil, err
			}
		}
		// update
		r.cur += 1
	}

	// next offset points to a next header
	r.next = r.cur + size

	h := &Header{
		Size:     size,
		Name:     name,
		ModTime:  modTime,
		GID:      gid,
		UID:      uid,
		Mode:     perm,
		nameSize: nameSize,
	}
	return h, nil
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
