//go:build !(js && wasm)

package platform

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCache(t *testing.T) {
	t.Parallel()

	c := &Cache{dir: t.TempDir()}

	if _, ok, err := c.Get("invaders"); ok || err != nil {
		t.Fatalf("empty cache: ok=%v err=%v", ok, err)
	}

	if err := c.Put("invaders", []byte{1, 2, 3}); err != nil {
		t.Fatal(err)
	}

	got, ok, err := c.Get("invaders")
	if !ok || err != nil || !bytes.Equal(got, []byte{1, 2, 3}) {
		t.Fatalf("Get = %v %v %v", got, ok, err)
	}

	if err := c.Delete("invaders"); err != nil {
		t.Fatal(err)
	}

	if _, ok, _ := c.Get("invaders"); ok {
		t.Error("still there after Delete")
	}

	if err := c.Delete("invaders"); err != nil {
		t.Errorf("deleting a missing key: %v", err)
	}
}

func TestCacheErrors(t *testing.T) {
	t.Parallel()

	c := &Cache{dir: t.TempDir()}
	for _, key := range []string{"", "../x", "a/b", "with space"} {
		if _, _, err := c.Get(key); !errors.Is(err, ErrBadKey) {
			t.Errorf("Get(%q): %v, want ErrBadKey", key, err)
		}

		if err := c.Put(key, nil); !errors.Is(err, ErrBadKey) {
			t.Errorf("Put(%q): %v, want ErrBadKey", key, err)
		}

		if err := c.Delete(key); !errors.Is(err, ErrBadKey) {
			t.Errorf("Delete(%q): %v, want ErrBadKey", key, err)
		}
	}

	none := &Cache{}
	if _, _, err := none.Get("k"); !errors.Is(err, ErrUnavailable) {
		t.Errorf("no folder: Get %v", err)
	}

	if err := none.Put("k", nil); !errors.Is(err, ErrUnavailable) {
		t.Errorf("no folder: Put %v", err)
	}

	if err := none.Delete("k"); !errors.Is(err, ErrUnavailable) {
		t.Errorf("no folder: Delete %v", err)
	}

	if NewCache() == nil {
		t.Error("NewCache returned nil")
	}
}

func TestPrefs(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "sub", "prefs.json")

	p := newPrefs(path)
	p.Set("volume", "30")
	p.Set("muted", "1")
	p.Set("muted", "") // deletes

	if !p.Persistent() {
		t.Error("not persistent with a writable file")
	}

	again := newPrefs(path)
	if v, ok := again.Get("volume"); !ok || v != "30" {
		t.Errorf("volume after reload = %q %v, want 30", v, ok)
	}

	if _, ok := again.Get("muted"); ok {
		t.Error("deleted key came back")
	}

	memory := newPrefs("")
	memory.Set("k", "v")

	if v, _ := memory.Get("k"); v != "v" || memory.Persistent() {
		t.Errorf("in-memory prefs: %q, persistent %v", v, memory.Persistent())
	}

	if err := os.WriteFile(path, []byte("{broken"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, ok := newPrefs(path).Get("volume"); ok {
		t.Error("a broken file should read as empty")
	}
}

func TestDesktopStubs(t *testing.T) {
	t.Parallel()

	if p := URLParams(); p.Volume != -1 {
		t.Errorf("URLParams = %+v, want none", p)
	}

	if CanPickFiles() {
		t.Error("CanPickFiles on the desktop")
	}

	if r := <-PickFiles(); !errors.Is(r.Err, ErrUnsupported) {
		t.Errorf("PickFiles: %v, want ErrUnsupported", r.Err)
	}

	if !InFront(true) || InFront(false) {
		t.Error("InFront should follow the window's focus")
	}

	PublishState(`{}`) // does nothing
}
