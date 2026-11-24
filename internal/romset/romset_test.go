package romset_test

import (
	"archive/zip"
	"bytes"
	"crypto/sha1" //nolint:gosec // G505: matches the package under test
	"encoding/hex"
	"errors"
	"hash/crc32"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/rbscholtus/Space-Invaders/internal/romset"
)

// chip returns 16 bytes of made-up ROM contents, different per seed.
func chip(seed byte) []byte {
	b := make([]byte, 16)
	for i := range b {
		b[i] = seed + byte(i*7)
	}

	return b
}

var (
	chipA = chip(1)
	chipB = chip(2)
)

func describe(name string, addr uint16, data []byte) romset.File {
	sum := sha1.Sum(data) //nolint:gosec // G401: as in the package

	return romset.File{
		Name: name, MAMEName: "chip_" + name, Addr: addr, Size: len(data),
		CRC: crc32.ChecksumIEEE(data), SHA1: hex.EncodeToString(sum[:]),
	}
}

// fakeSet is a two-chip set made from chipA and chipB, so the tests need no
// real ROMs.
func fakeSet() romset.Set {
	return romset.Set{
		Name: "fake", ImageSize: 32,
		Files: []romset.File{describe("fake.a", 0, chipA), describe("fake.b", 16, chipB)},
	}
}

func wantImage() []byte { return append(append([]byte{}, chipA...), chipB...) }

func TestLoad(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		files   fstest.MapFS
		wantErr error
		wantMsg string // substring of the error
	}{
		{name: "by name", files: fstest.MapFS{"fake.a": {Data: chipA}, "fake.b": {Data: chipB}}},
		{name: "renamed", files: fstest.MapFS{"x": {Data: chipB}, "y.bin": {Data: chipA}}},
		{name: "in a folder, with extras", files: fstest.MapFS{"set/fake.a": {Data: chipA}, "set/fake.b": {Data: chipB}, "readme.txt": {Data: []byte("hi")}}},
		{name: "one missing", files: fstest.MapFS{"fake.a": {Data: chipA}}, wantErr: romset.ErrMissingROM, wantMsg: "fake.b"},
		{name: "none", files: fstest.MapFS{"other.rom": {Data: chip(9)}}, wantErr: romset.ErrNotAROMSet, wantMsg: "found: other.rom"},
		{name: "empty", files: fstest.MapFS{}, wantErr: romset.ErrNotAROMSet, wantMsg: "no files"},
		{name: "wrong contents", files: fstest.MapFS{"fake.a": {Data: chipA}, "FAKE.B": {Data: chip(9)}}, wantErr: romset.ErrCRCMismatch, wantMsg: "FAKE.B has CRC"},
		{name: "wrong contents, MAME name", files: fstest.MapFS{"fake.a": {Data: chipA}, "chip_fake.b": {Data: chip(9)}}, wantErr: romset.ErrCRCMismatch},
		{name: "too big to be a ROM", files: fstest.MapFS{"big.bin": {Data: make([]byte, 100<<10)}}, wantErr: romset.ErrNotAROMSet, wantMsg: "big.bin"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			image, err := romset.Load(fakeSet(), tt.files)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatal(err)
				}

				if !bytes.Equal(image, wantImage()) {
					t.Errorf("image % X, want % X", image, wantImage())
				}

				return
			}

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}

			if !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("err = %q, want it to mention %q", err, tt.wantMsg)
			}
		})
	}
}

// TestSHA1Mismatch: a file whose CRC matches but whose SHA-1 doesn't is
// rejected (a CRC collision, or a wrong table entry).
func TestSHA1Mismatch(t *testing.T) {
	t.Parallel()

	set := fakeSet()
	set.Files[1].SHA1 = strings.Repeat("0", 40)

	_, err := romset.Load(set, fstest.MapFS{"fake.a": {Data: chipA}, "fake.b": {Data: chipB}})
	if !errors.Is(err, romset.ErrCRCMismatch) || !strings.Contains(err.Error(), "SHA-1") {
		t.Errorf("err = %v, want a SHA-1 mismatch", err)
	}
}

func zipOf(t *testing.T, files map[string][]byte) []byte {
	t.Helper()

	var buf bytes.Buffer

	zw := zip.NewWriter(&buf)
	for name, data := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := w.Write(data); err != nil {
			t.Fatal(err)
		}
	}

	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}

func TestLoadZip(t *testing.T) {
	t.Parallel()

	image, err := romset.LoadZip(fakeSet(), zipOf(t, map[string][]byte{"fake.a": chipA, "dir/renamed": chipB}))
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(image, wantImage()) {
		t.Errorf("image % X, want % X", image, wantImage())
	}

	if _, err := romset.LoadZip(fakeSet(), []byte("not a zip")); err == nil {
		t.Error("LoadZip accepted something that isn't a zip")
	}
}

// TestInvadersTable checks the set's layout: four 2 KB chips that fill the
// 8 KB image without gaps or overlaps.
func TestInvadersTable(t *testing.T) {
	t.Parallel()

	set := romset.Invaders()
	next := 0

	for _, f := range set.Files {
		if int(f.Addr) != next || f.Size != 0x800 || len(f.SHA1) != 40 {
			t.Errorf("%s: addr %04X size %d SHA-1 %q", f.Name, f.Addr, f.Size, f.SHA1)
		}

		next += f.Size
	}

	if next != set.ImageSize {
		t.Errorf("files end at %04X, image is %04X", next, set.ImageSize)
	}
}

// TestRealROMs loads your own ROM set (DESIGN §15.3): roms/invaders.zip, or
// the zip or folder in INVADERS_ROMS. Skipped without one; CI has none.
func TestRealROMs(t *testing.T) {
	t.Parallel()

	path := os.Getenv("INVADERS_ROMS")
	if path == "" {
		path = filepath.Join("..", "..", "roms", "invaders.zip")
	}

	info, err := os.Stat(path) //nolint:gosec // G703: the developer's own ROM path, from their environment
	if err != nil {
		t.Skipf("no ROM set at %s (set INVADERS_ROMS to use another)", path)
	}

	var image []byte

	if info.IsDir() {
		image, err = romset.Load(romset.Invaders(), os.DirFS(path))
	} else {
		var data []byte

		data, err = os.ReadFile(path) //nolint:gosec // G703: as above
		if err == nil {
			image, err = romset.LoadZip(romset.Invaders(), data)
		}
	}

	if err != nil {
		t.Fatal(err)
	}

	if len(image) != 0x2000 {
		t.Errorf("image is %d bytes, want 8192", len(image))
	}
}
