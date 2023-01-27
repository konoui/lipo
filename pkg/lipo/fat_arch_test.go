// Note internal package test
package lipo

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_createTemp(t *testing.T) {
	tests := []struct {
		name     string
		dir      string
		filename string
		wantErr  bool
	}{
		{name: "abs", dir: os.TempDir(), filename: "output"},
		{name: "rela", dir: "./", filename: "output"},
		{name: "rela", dir: "", filename: "output"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tt.dir, tt.filename)
			f, err := createTemp(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("createTemp() error = %v, wantErr %v\n", err, tt.wantErr)
				return
			}

			t.Cleanup(func() { os.RemoveAll(f.Name()) })

			want := filepath.Clean(tt.dir)
			got := filepath.Dir(f.Name())
			if want != got {
				t.Errorf("createTemp() = %v, want %v\n", got, want)
				t.Logf("tmp file: %s, input dir: %s\n", f.Name(), tt.dir)
			}
		})
	}
}
