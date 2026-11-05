package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/mod/semver"
)

const (
	LatestReleaseURL = "https://api.github.com/repos/phanivarun05/yamlchecker/releases/latest"
	cacheTTL         = 24 * time.Hour
	failureTTL       = 1 * time.Hour
)

// Release is what the rest of the program needs to know about a version.
type Release struct {
	Tag string
	URL string
}

// cacheEntry is what we store on disk between runs.
type cacheEntry struct {
	CheckedAt time.Time `json:"checked_at"`
	Tag       string    `json:"tag"`
	URL       string    `json:"url"`
	Failed    bool      `json:"failed,omitempty"`
}

// Checker holds everything that differs between production and tests.
type Checker struct {
	Client    *http.Client
	URL       string
	CacheFile string
}

// IsNewer reports whether latest is a higher version than current.
// If either one isn't valid semver (like "dev"), the answer is false.
func IsNewer(current, latest string) bool {
	if !semver.IsValid(current) || !semver.IsValid(latest) {
		return false
	}
	return semver.Compare(latest, current) > 0
}

// IsReleaseVersion is true only for clean releases like "v0.1.2".
// It rejects "dev", pseudo-versions and anything with a suffix.
func IsReleaseVersion(v string) bool {
	return semver.IsValid(v) && semver.Prerelease(v) == "" && semver.Build(v) == ""
}

// Eligible decides whether this run should check for updates at all.
func Eligible(version string) bool {
	if !IsReleaseVersion(version) {
		return false
	}
	if os.Getenv("YAMLCHECKER_NO_UPDATE_CHECK") != "" {
		return false
	}
	if os.Getenv("CI") != "" {
		return false
	}
	return stderrIsTerminal()
}

func stderrIsTerminal() bool {
	info, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// DefaultCacheFile returns the per-user location of the cache file.
func DefaultCacheFile() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "yamlchecker", "update-check.json"), nil
}

// Latest returns the newest release, using the cache when it is fresh.
func (c *Checker) Latest(ctx context.Context) (Release, error) {
	cached, haveCache := c.readCache()
	if haveCache && cached.fresh() {
		if cached.Tag == "" {
			return Release{}, errors.New("response had no tag_name")
		}
		return Release{Tag: cached.Tag, URL: cached.URL}, nil
	}

	rel, err := c.fetch(ctx)
	if err != nil {
		// Remember that we tried, so an offline machine doesn't
		// wait for a timeout on every single run.
		c.writeCache(cacheEntry{CheckedAt: time.Now(), Tag: cached.Tag, URL: cached.URL})
		return Release{}, err
	}

	c.writeCache(cacheEntry{CheckedAt: time.Now(), Tag: rel.Tag, URL: rel.URL})
	return rel, nil
}

func (c *Checker) fetch(ctx context.Context) (Release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL, nil)
	if err != nil {
		return Release{}, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("User-Agent", "yamlchecker-update-check") // GitHub rejects requests without one
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return Release{}, fmt.Errorf("requesting %s: %w", c.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("unexpected status %s", resp.Status)
	}

	var payload struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	// LimitReader caps the body at 1 MB, so a misbehaving server can't flood memory.
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return Release{}, fmt.Errorf("decoding response: %w", err)
	}
	if payload.TagName == "" {
		return Release{}, errors.New("response had no tag_name")
	}
	return Release{Tag: payload.TagName, URL: payload.HTMLURL}, nil
}

func (c *Checker) readCache() (cacheEntry, bool) {
	data, err := os.ReadFile(c.CacheFile)
	if err != nil {
		return cacheEntry{}, false
	}
	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return cacheEntry{}, false // a corrupt file counts as "no cache"
	}
	return entry, true
}

// writeCache is best-effort: if it fails, the next run just checks again.
func (c *Checker) writeCache(entry cacheEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(c.CacheFile), 0o755); err != nil {
		return
	}
	_ = os.WriteFile(c.CacheFile, data, 0o644)
}

// Cached returns the stored release without checking the network.
func (c *Checker) Cached() (Release, bool) {
	entry, ok := c.readCache()
	if !ok || entry.Tag == "" {
		return Release{}, false
	}
	return Release{Tag: entry.Tag, URL: entry.URL}, true
}

// Stale reports whether the cached release is older than cacheTTL.
func (c *Checker) Stale() bool {
	entry, ok := c.readCache()
	return !ok || !entry.fresh()
}

func (e cacheEntry) fresh() bool {
	ttl := cacheTTL
	if e.Failed || e.Tag == "" {
		ttl = failureTTL
	}
	return time.Since(e.CheckedAt) < ttl
}
