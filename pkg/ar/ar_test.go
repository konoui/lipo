package ar_test

import (
	"bytes"
	"crypto/sha256"
	"debug/macho"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/konoui/lipo/pkg/ar"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     []string
	}{
		{
			name:     "arm64-func1.a",
			filename: "arm64-func1.a",
			want:     []string{"arm64-func1.o"},
		},
		{
			name:     "arm64-func12.a",
			filename: "arm64-func12.a",
			want:     []string{"arm64-func1.o", "arm64-func2.o"},
		},
		{
			name:     "arm64-func123.a",
			filename: "arm64-func123.a",
			want:     []string{"arm64-func1.o", "arm64-func2.o", "arm64-func3.o"},
		},
		{
			name:     "arm64-amd64-func12.a",
			filename: "arm64-amd64-func12.a",
			want:     []string{"amd64-func1.o", "amd64-func2.o", "arm64-func1.o", "arm64-func2.o"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := filepath.Join("testdata", tt.filename)
			f, err := os.Open(p)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			archiver, err := ar.NewReader(f)
			if err != nil {
				t.Fatal(err)
			}

			got := []string{}
			for {
				file, err := archiver.Next()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					t.Fatal(err)
				}
				if strings.HasPrefix(file.Name, ar.PrefixSymdef) {
					continue
				}

				if tt.filename == "arm64-func1.a" && file.Name == "arm64-func1.o" {
					if file.UID != 501 {
						t.Errorf("uid: want %d, got %d\n", 501, file.UID)
					}
					if file.GID != 20 {
						t.Errorf("gid: want %d, got %d\n", 20, file.GID)
					}
					if file.ModTime.Unix() != 1697716653 {
						t.Errorf("modtime: want %v, got %v\n", 1697716653, file.ModTime.Unix())
					}
					if file.Mode.Perm() != fs.FileMode(0644) {
						t.Errorf("mode: want %d, got %d\n", fs.FileMode(0644), file.Mode.Perm())
					}
				}

				got = append(got, file.Name)

				machoBuf := &bytes.Buffer{}
				tee := io.TeeReader(file, machoBuf)
				gotHash := sha256String(t, tee)
				wantHash := sha256StringFromFile(t, filepath.Join("testdata", file.Name))
				if gotHash != wantHash {
					t.Errorf("%s: want: %s got: %s", file.Name, wantHash, gotHash)
				}

				// validate
				if _, err := macho.NewFile(bytes.NewReader(machoBuf.Bytes())); err != nil {
					t.Fatal(err)
				}
			}

			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("want: %v, got: %v", tt.want, got)
			}
		})
	}
}

func sha256StringFromFile(t *testing.T, p string) string {
	f, err := os.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	return sha256String(t, f)
}

func sha256String(t *testing.T, r io.Reader) string {
	h := sha256.New()
	_, err := io.Copy(h, r)
	if err != nil {
		t.Fatal(err)
	}

	return hex.EncodeToString(h.Sum(nil))
}

func Test_TrimTailSpace(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{
			name:  "space1",
			input: []byte(" test"),
			want:  " test",
		},
		{
			name:  "space2",
			input: []byte(" test "),
			want:  " test",
		},
		{
			name:  "space3",
			input: []byte("test test "),
			want:  "test test",
		},
		{
			name:  "space4",
			input: []byte(" test test "),
			want:  " test test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ar.TrimTailSpace(tt.input); got != tt.want {
				t.Errorf("TrimTailSpace() = [%v:, want [%v]", got, tt.want)
			}
		})
	}
}
