// Package romset finds, checks and assembles a game's ROM set from the files a
// player brings: a MAME zip or the loose files (DESIGN §9). Files are matched
// by size and CRC-32, so renamed files work, and then checked by SHA-1.
//
// It uses only the standard library and never fetches anything.
package romset

import (
	"archive/zip"
	"bytes"
	"crypto/sha1" //nolint:gosec // G505: SHA-1 identifies ROM dumps (as MAME does); not used for security
	"encoding/hex"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"io/fs"
	"slices"
	"strings"
)

var (
	// ErrNotAROMSet means none of the files belong to the set.
	ErrNotAROMSet = errors.New("no ROMs of this set found")
	// ErrMissingROM means some, but not all, of the set's files were found.
	ErrMissingROM = errors.New("ROM missing")
	// ErrCRCMismatch means a file has the expected name but other contents:
	// a different version, a bootleg or a bad dump.
	ErrCRCMismatch = errors.New("CRC mismatch")
)

// Size limits, so a large or hostile input can't exhaust memory: ROM chips
// are a few KB, and a zip found among the files (a dropped or picked
// invaders.zip) is opened only up to maxZipSize.
const (
	maxFileSize = 64 << 10
	maxZipSize  = 16 << 20
)

// zipMagic starts every zip file (a local file header).
const zipMagic = "PK\x03\x04"

// File is one ROM chip of a set.
type File struct {
	Name     string // as in older MAME sets, e.g. "invaders.h"
	MAMEName string // as in current MAME, e.g. "9316b-0869_m739h.h1"
	Addr     uint16 // where it goes in the image
	Size     int
	CRC      uint32
	SHA1     string // lower-case hex
}

// Set is a ROM set: the files that make up one game's program image.
type Set struct {
	Name      string // the MAME set name, e.g. "invaders"
	ImageSize int    // the size of the assembled image
	Files     []File
}

// Invaders is MAME's "invaders" set: Space Invaders, Midway, 1978
// (DESIGN §9.1; MAME mw8080bw.cpp, ROM_START( invaders )).
func Invaders() Set {
	return Set{
		Name:      "invaders",
		ImageSize: 0x2000,
		Files: []File{
			{"invaders.h", "9316b-0869_m739h.h1", 0x0000, 0x800, 0x734f5ad8, "ff6200af4c9110d8181249cbcef1a8a40fa40b7f"},
			{"invaders.g", "9316b-0856_m739g.g1", 0x0800, 0x800, 0x6bfaca4a, "16f48649b531bdef8c2d1446c429b5f414524350"},
			{"invaders.f", "9316b-0855_m739f.f1", 0x1000, 0x800, 0x0ccead96, "537aef03468f63c5b9e11dd61e253f7ae17d9743"},
			{"invaders.e", "9316b-0854_m739e.e1", 0x1800, 0x800, 0x14e538b0, "1d6ca0c99f9df71e2990b610deb9d7da0125e2d8"},
		},
	}
}

// LoadZip assembles set from a zip archive in memory.
func LoadZip(set Set, data []byte) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("reading the zip: %w", err)
	}

	return Load(set, zr)
}

// candidate is a file found in the input.
type candidate struct {
	path string
	data []byte
	crc  uint32
}

// Load assembles set from the files in fsys, searched recursively (a folder
// via os.DirFS, or a zip.Reader). Each of the set's files is found by its
// size and CRC-32, whatever its name, and checked by SHA-1.
func Load(set Set, fsys fs.FS) ([]byte, error) {
	found, err := scan(fsys)
	if err != nil {
		return nil, err
	}

	image := make([]byte, set.ImageSize)

	var missing, mismatched []string

	matched := 0

	for _, f := range set.Files {
		c := match(f, found)
		if c == nil {
			if wrong := byName(f, found); wrong != nil {
				mismatched = append(mismatched,
					fmt.Sprintf("%s has CRC %08x, want %08x", wrong.path, wrong.crc, f.CRC))
			} else {
				missing = append(missing, f.Name)
			}

			continue
		}

		sum := sha1.Sum(c.data) //nolint:gosec // G401: identifies the dump, see the import
		if got := hex.EncodeToString(sum[:]); got != f.SHA1 {
			mismatched = append(mismatched,
				fmt.Sprintf("%s matches %s's CRC but has SHA-1 %s, want %s", c.path, f.Name, got, f.SHA1))

			continue
		}

		copy(image[f.Addr:], c.data)

		matched++
	}

	switch {
	case len(mismatched) > 0:
		return nil, fmt.Errorf("%w: %s (a different version, a bootleg or a bad dump?)",
			ErrCRCMismatch, strings.Join(mismatched, "; "))
	case matched == 0:
		return nil, fmt.Errorf("%w (%s); found: %s", ErrNotAROMSet, set.Name, names(found))
	case len(missing) > 0:
		return nil, fmt.Errorf("%w: %s", ErrMissingROM, strings.Join(missing, ", "))
	}

	return image, nil
}

// scan reads every regular file in fsys up to maxFileSize; larger files
// can't be ROM chips and are only listed. A zip among the files is opened
// and its entries scanned too (one level), so a dropped invaders.zip works
// like the files inside it.
func scan(fsys fs.FS) ([]candidate, error) {
	return scanFS(fsys, "", true)
}

func scanFS(fsys fs.FS, prefix string, openZips bool) ([]candidate, error) {
	var found []candidate

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.Type().IsRegular() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}

		limit := int64(maxFileSize)
		if openZips {
			limit = maxZipSize
		}

		if info.Size() > limit {
			found = append(found, candidate{path: prefix + path})

			return nil
		}

		data, err := readFile(fsys, path, limit)
		if err != nil {
			return err
		}

		if openZips && strings.HasPrefix(string(data[:min(len(data), len(zipMagic))]), zipMagic) {
			inner, err := scanZip(data, prefix+path+"/")
			if err != nil {
				return err
			}

			found = append(found, inner...)

			return nil
		}

		if len(data) > maxFileSize {
			found = append(found, candidate{path: prefix + path})

			return nil
		}

		found = append(found, candidate{path: prefix + path, data: data, crc: crc32.ChecksumIEEE(data)})

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("reading the ROM files: %w", err)
	}

	return found, nil
}

// scanZip scans the entries of a zip found among the files.
func scanZip(data []byte, prefix string) ([]candidate, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", strings.TrimSuffix(prefix, "/"), err)
	}

	return scanFS(zr, prefix, false)
}

// readFile reads at most limit bytes, whatever size the entry claims.
func readFile(fsys fs.FS, path string, limit int64) ([]byte, error) {
	f, err := fsys.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, limit))
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	return data, nil
}

// match returns the candidate with f's size and CRC, or nil.
func match(f File, found []candidate) *candidate {
	for i := range found {
		c := &found[i]
		if len(c.data) == f.Size && c.crc == f.CRC {
			return c
		}
	}

	return nil
}

// byName returns the candidate named like f (either name, any folder, any
// case), or nil.
func byName(f File, found []candidate) *candidate {
	for i := range found {
		base := found[i].path[strings.LastIndex(found[i].path, "/")+1:]
		if strings.EqualFold(base, f.Name) || strings.EqualFold(base, f.MAMEName) {
			return &found[i]
		}
	}

	return nil
}

func names(found []candidate) string {
	if len(found) == 0 {
		return "no files"
	}

	list := make([]string, len(found))
	for i, c := range found {
		list[i] = c.path
	}

	slices.Sort(list)

	return strings.Join(list, ", ")
}

// ErrBadImage is returned by CheckImage for an image that isn't the set.
var ErrBadImage = errors.New("not this ROM set's image")

// CheckImage checks an assembled image (for example one remembered from an
// earlier visit) against the set: its size, and each chip's CRC-32 and
// SHA-1.
func CheckImage(set Set, image []byte) error {
	if len(image) != set.ImageSize {
		return fmt.Errorf("%w: %d bytes, want %d", ErrBadImage, len(image), set.ImageSize)
	}

	for _, f := range set.Files {
		chip := image[f.Addr : int(f.Addr)+f.Size]
		sum := sha1.Sum(chip) //nolint:gosec // G401: identifies the dump, see the import

		if crc32.ChecksumIEEE(chip) != f.CRC || hex.EncodeToString(sum[:]) != f.SHA1 {
			return fmt.Errorf("%w: %s differs", ErrBadImage, f.Name)
		}
	}

	return nil
}
