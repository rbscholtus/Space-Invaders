package platform

import "sync"

// Prefs are small string settings (volume, mute). In the browser they live
// in localStorage, on the desktop in a JSON file in the user's config
// folder. Reads and writes are quick and never fail: when storage is
// blocked (private browsing, disabled cookies) the values live in memory
// for this session only.
type Prefs struct {
	path   string // desktop only: the preferences file
	mu     sync.Mutex
	memory map[string]string
	saved  bool // false once storage has failed
}

// NewPrefs returns the preferences store.
func NewPrefs() *Prefs { return newPrefs(defaultPrefsPath()) }

func newPrefs(path string) *Prefs {
	p := &Prefs{path: path, memory: map[string]string{}, saved: true}
	p.load()

	return p
}

// Get returns a preference.
func (p *Prefs) Get(key string) (string, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	v, ok := p.memory[key]

	return v, ok
}

// Set stores a preference; an empty value deletes it.
func (p *Prefs) Set(key, value string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if value == "" {
		delete(p.memory, key)
	} else {
		p.memory[key] = value
	}

	if p.saved && !p.store(key, value) {
		p.saved = false
	}
}

// Persistent reports whether preferences survive a restart.
func (p *Prefs) Persistent() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.saved
}
