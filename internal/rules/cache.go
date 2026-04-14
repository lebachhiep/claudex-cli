package rules

import (
	"os"
	"path/filepath"
	"sort"
)

const maxCacheItems = 10

// Cache manages local version bundles at ~/.claudex/cache/.
type Cache struct {
	Dir string
}

// NewCache creates a cache backed by the given directory.
func NewCache(cacheDir string) *Cache {
	return &Cache{Dir: cacheDir}
}

func (c *Cache) path(version string) string {
	return filepath.Join(c.Dir, version+".zip")
}

// GetIfValid reads cached bundle if it exists and checksum matches.
// Returns nil, nil on cache miss (file missing or checksum mismatch).
func (c *Cache) GetIfValid(version, checksum string) ([]byte, error) {
	data, err := os.ReadFile(c.path(version))
	if err != nil {
		return nil, nil // cache miss
	}
	if VerifyChecksum(data, checksum) != nil {
		_ = os.Remove(c.path(version)) // stale entry
		return nil, nil                // cache miss
	}
	return data, nil
}

// Put writes bundle to cache and evicts oldest entries over maxCacheItems.
func (c *Cache) Put(version string, data []byte) error {
	if err := os.MkdirAll(c.Dir, 0700); err != nil {
		return err
	}
	if err := os.WriteFile(c.path(version), data, 0600); err != nil {
		return err
	}
	return c.evict()
}

// evict removes the oldest cache entries when count exceeds maxCacheItems.
func (c *Cache) evict() error {
	entries, err := os.ReadDir(c.Dir)
	if err != nil {
		return nil
	}

	// Filter only .zip files
	var zips []os.DirEntry
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".zip" {
			zips = append(zips, e)
		}
	}

	if len(zips) <= maxCacheItems {
		return nil
	}

	// Sort by modification time (oldest first)
	sort.Slice(zips, func(i, j int) bool {
		fi, _ := zips[i].Info()
		fj, _ := zips[j].Info()
		if fi == nil || fj == nil {
			return false
		}
		return fi.ModTime().Before(fj.ModTime())
	})

	// Remove oldest
	toRemove := len(zips) - maxCacheItems
	for i := 0; i < toRemove; i++ {
		_ = os.Remove(filepath.Join(c.Dir, zips[i].Name()))
	}
	return nil
}
