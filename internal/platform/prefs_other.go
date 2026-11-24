//go:build !(js && wasm)

package platform

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// defaultPrefsPath is the app's file in the user config folder, or "" if
// there is none (then preferences live in memory only).
func defaultPrefsPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}

	return filepath.Join(dir, appName, "prefs.json")
}

// load reads the preferences file; a missing or broken file means no
// preferences yet.
func (p *Prefs) load() {
	if p.path == "" {
		p.saved = false

		return
	}

	data, err := os.ReadFile(p.path)
	if err != nil {
		return
	}

	_ = json.Unmarshal(data, &p.memory)
}

// store writes all preferences (the file is small). The lock is held.
func (p *Prefs) store(_, _ string) bool {
	if p.path == "" {
		return false
	}

	data, err := json.MarshalIndent(p.memory, "", "  ")
	if err != nil {
		return false
	}

	if err := os.MkdirAll(filepath.Dir(p.path), 0o750); err != nil {
		return false
	}

	return os.WriteFile(p.path, data, 0o600) == nil
}
