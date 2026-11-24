package platform

// Cache keeps the player's ROM image between visits (DESIGN §9.2): in the
// browser in IndexedDB, on the desktop in the user cache folder. Its calls
// may wait on I/O: use them from a goroutine, never from Update or Draw.
type Cache struct {
	dir string // desktop only: the folder
}

// storeROMs is the IndexedDB object store (or sub-folder) for ROM images.
const storeROMs = "roms"

// NewCache returns the ROM cache.
func NewCache() *Cache { return &Cache{dir: defaultCacheDir()} }

// Get returns the value for key; the bool is false if there is none.
func (c *Cache) Get(key string) ([]byte, bool, error) {
	if !validKey.MatchString(key) {
		return nil, false, ErrBadKey
	}

	return c.get(key)
}

// Put stores value under key.
func (c *Cache) Put(key string, value []byte) error {
	if !validKey.MatchString(key) {
		return ErrBadKey
	}

	return c.put(key, value)
}

// Delete removes key; a missing key is not an error.
func (c *Cache) Delete(key string) error {
	if !validKey.MatchString(key) {
		return ErrBadKey
	}

	return c.delete(key)
}
