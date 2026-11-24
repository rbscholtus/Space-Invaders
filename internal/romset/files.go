package romset

import (
	"bytes"
	"io"
	"io/fs"
	"slices"
	"time"
)

// Files returns the given files (name → contents) as an fs.FS, so files
// picked in the browser load like a folder or a drop. Names are taken as
// they are, without directories.
func Files(files map[string][]byte) fs.FS { return memFS(files) }

type memFS map[string][]byte

// Open implements fs.FS.
func (m memFS) Open(name string) (fs.File, error) {
	if name == "." {
		return &memDir{fs: m}, nil
	}

	data, ok := m[name]
	if !ok || !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}

	return &memFile{info: memInfo{name: name, size: int64(len(data))}, r: bytes.NewReader(data)}, nil
}

// ReadDir implements fs.ReadDirFS: all files are in the root.
func (m memFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if name != "." {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
	}

	names := make([]string, 0, len(m))
	for n := range m {
		if fs.ValidPath(n) {
			names = append(names, n)
		}
	}

	slices.Sort(names)

	entries := make([]fs.DirEntry, len(names))
	for i, n := range names {
		entries[i] = fs.FileInfoToDirEntry(memInfo{name: n, size: int64(len(m[n]))})
	}

	return entries, nil
}

type memFile struct {
	info memInfo
	r    *bytes.Reader
}

func (f *memFile) Stat() (fs.FileInfo, error) { return f.info, nil }

// Read passes the reader's io.EOF through unwrapped, as io.Reader requires.
func (f *memFile) Read(p []byte) (int, error) {
	return f.r.Read(p) //nolint:wrapcheck // io.EOF must not be wrapped
}
func (f *memFile) Close() error { return nil }

// memDir is the root directory, opened with Open(".").
type memDir struct {
	fs   memFS
	read int // entries already returned by ReadDir
}

func (d *memDir) Stat() (fs.FileInfo, error) { return memInfo{name: ".", dir: true}, nil }
func (d *memDir) Read([]byte) (int, error)   { return 0, io.EOF }
func (d *memDir) Close() error               { return nil }

// ReadDir implements fs.ReadDirFile: n <= 0 returns all remaining entries,
// otherwise at most n, and io.EOF at the end.
func (d *memDir) ReadDir(n int) ([]fs.DirEntry, error) {
	all, _ := d.fs.ReadDir(".")
	rest := all[d.read:]

	if n <= 0 {
		d.read = len(all)

		return rest, nil
	}

	if len(rest) == 0 {
		return nil, io.EOF
	}

	rest = rest[:min(n, len(rest))]
	d.read += len(rest)

	return rest, nil
}

type memInfo struct {
	name string
	size int64
	dir  bool
}

func (i memInfo) Name() string { return i.name }
func (i memInfo) Size() int64  { return i.size }
func (i memInfo) Mode() fs.FileMode {
	if i.dir {
		return fs.ModeDir | 0o555
	}

	return 0o444
}
func (i memInfo) ModTime() time.Time { return time.Time{} }
func (i memInfo) IsDir() bool        { return i.dir }
func (i memInfo) Sys() any           { return nil }
