package ar_test

import (
	"bytes"
	"crypto/sha256"
	"debug/macho"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/konoui/lipo/pkg/lipo/ar"
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

			archiver, err := ar.New(f)
			if err != nil {
				t.Fatal(err)
			}

			got := []string{}
			for {
				file, err := archiver.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatal(err)
				}
				if strings.HasPrefix(file.Name, ar.PrefixSymdef) {
					continue
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
