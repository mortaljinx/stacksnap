package snapshot

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pmezard/go-difflib/difflib"
)

// Store manages versioned compose snapshots on disk.
type Store struct {
	baseDir string
	keep    int
}

// NewStore creates a Store rooted at baseDir, retaining up to keep versions.
func NewStore(baseDir string, keep int) *Store {
	return &Store{baseDir: baseDir, keep: keep}
}

// Save snapshots yamlData for the given stack name.
// It is a no-op if the content has not changed since the last snapshot.
func (s *Store) Save(name string, yamlData []byte) error {
	if len(yamlData) == 0 {
		return fmt.Errorf("refusing to save empty compose data for %q", name)
	}

	dir := filepath.Join(s.baseDir, sanitiseName(name))
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("create snapshot dir: %w", err)
	}

	// ── Change detection ────────────────────────────────────────────────────
	newHash := hashOf(yamlData)
	hashFile := filepath.Join(dir, ".hash")

	oldHash, _ := os.ReadFile(hashFile)
	if strings.TrimSpace(string(oldHash)) == newHash {
		return nil // unchanged — nothing to do
	}

	// ── Load previous version for diff ─────────────────────────────────────
	latestFile := filepath.Join(dir, "latest.yml")
	oldData, _ := os.ReadFile(latestFile)

	// ── Write new version ───────────────────────────────────────────────────
	timestamp := time.Now().Format("2006-01-02_1504")
	versionFile := filepath.Join(dir, timestamp+".yml")

	if err := writeFile(versionFile, yamlData); err != nil {
		return fmt.Errorf("write version file: %w", err)
	}
	if err := writeFile(latestFile, yamlData); err != nil {
		return fmt.Errorf("write latest.yml: %w", err)
	}

	// ── Write diff (only if there is a previous version to diff against) ───
	if len(oldData) > 0 {
		diff := difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(oldData)),
			B:        difflib.SplitLines(string(yamlData)),
			FromFile: "previous",
			ToFile:   "current",
			Context:  3,
		}
		diffText, err := difflib.GetUnifiedDiffString(diff)
		if err != nil {
			return fmt.Errorf("generate diff: %w", err)
		}
		if diffText != "" {
			diffFile := filepath.Join(dir, timestamp+".diff")
			if err := writeFile(diffFile, []byte(diffText)); err != nil {
				return fmt.Errorf("write diff file: %w", err)
			}
		}
	}

	// ── Update hash ─────────────────────────────────────────────────────────
	if err := writeFile(hashFile, []byte(newHash)); err != nil {
		return fmt.Errorf("write hash file: %w", err)
	}

	// ── Rotate old versions ─────────────────────────────────────────────────
	if err := s.rotate(dir); err != nil {
		return fmt.Errorf("rotate old versions: %w", err)
	}

	return nil
}

// rotate removes the oldest timestamped versions beyond the keep limit.
// It removes both .yml and .diff files for each pruned version.
func (s *Store) rotate(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var versions []string
	for _, e := range entries {
		n := e.Name()
		if strings.HasSuffix(n, ".yml") && n != "latest.yml" {
			versions = append(versions, strings.TrimSuffix(n, ".yml"))
		}
	}

	sort.Strings(versions) // oldest first

	if len(versions) <= s.keep {
		return nil
	}

	for _, old := range versions[:len(versions)-s.keep] {
		os.Remove(filepath.Join(dir, old+".yml"))
		os.Remove(filepath.Join(dir, old+".diff"))
	}

	return nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func hashOf(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// writeFile writes data atomically-ish: write to a temp file then rename.
// This prevents leaving truncated files if the process is interrupted.
func writeFile(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0640); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// sanitiseName removes path-traversal characters from stack names.
func sanitiseName(name string) string {
	// Replace anything that isn't alphanumeric, dash, underscore, or dot.
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	s := b.String()
	// Reject directory traversal attempts.
	s = strings.Trim(s, ".")
	if s == "" {
		s = "unknown"
	}
	return s
}
