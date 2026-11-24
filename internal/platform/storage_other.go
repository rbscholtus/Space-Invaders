//go:build !(js && wasm)

package platform

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// defaultCacheDir is the app's folder in the user cache directory, or ""
// if there is none (then the cache is unavailable).
func defaultCacheDir() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}

	return filepath.Join(dir, appName)
}

func (c *Cache) path(key string) string { return filepath.Join(c.dir, storeROMs, key) }

func (c *Cache) get(key string) ([]byte, bool, error) {
	if c.dir == "" {
		return nil, false, ErrUnavailable
	}

	data, err := os.ReadFile(c.path(key))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, fmt.Errorf("cache get: %w", err)
	}

	return data, true, nil
}

func (c *Cache) put(key string, value []byte) error {
	if c.dir == "" {
		return ErrUnavailable
	}

	dir := filepath.Join(c.dir, storeROMs)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("cache put: %w", err)
	}

	// Write to a temporary file and rename, so a crash never leaves half a
	// value behind.
	tmp, err := os.CreateTemp(dir, key+".*")
	if err != nil {
		return fmt.Errorf("cache put: %w", err)
	}

	if _, err := tmp.Write(value); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())

		return fmt.Errorf("cache put: %w", err)
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())

		return fmt.Errorf("cache put: %w", err)
	}

	if err := os.Rename(tmp.Name(), c.path(key)); err != nil {
		return fmt.Errorf("cache put: %w", err)
	}

	return nil
}

func (c *Cache) delete(key string) error {
	if c.dir == "" {
		return ErrUnavailable
	}

	if err := os.Remove(c.path(key)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("cache delete: %w", err)
	}

	return nil
}
