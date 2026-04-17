package rules

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxFileSize = 100 * 1024 * 1024 // 100MB per extracted file

// InstallMode controls overwrite behavior.
type InstallMode int

const (
	ModeInit   InstallMode = iota // Full overwrite (first install)
	ModeUpdate                    // Selective overwrite (preserve user files)
)

// preservePaths lists paths within .claude/ that should be preserved during update.
var preservePaths = []string{
	".env",
	".claude-config.json",
	"session-state",
	"hooks/.logs",
}

// neverOverwrite lists root-level files that should not be overwritten if they already exist.
// These files belong to the project, not to claudex.
var neverOverwrite = []string{
	".gitignore",
}

// FileStats tracks what was extracted from the bundle.
type FileStats struct {
	SkillCount  int
	AgentCount  int
	RuleCount   int
	HasClaudeMD bool
}

// Install extracts the rules bundle into the target project directory.
func Install(result *DownloadResult, plan string, targetDir string, force bool, cliVersion string) (*FileStats, error) {
	return InstallWithMode(result, plan, targetDir, force, cliVersion, ModeInit)
}

// InstallWithMode extracts the rules bundle with configurable overwrite mode.
func InstallWithMode(result *DownloadResult, plan string, targetDir string, force bool, cliVersion string, mode InstallMode) (*FileStats, error) {
	claudeDir := filepath.Join(targetDir, ".claude")

	// Check if .claude/ already exists
	if _, err := os.Stat(claudeDir); err == nil && !force && mode == ModeInit {
		return nil, fmt.Errorf(".claude/ already exists in this project.\nHint: Use `claudex init --force` to overwrite")
	}

	// Check for symlink safety
	if info, err := os.Lstat(claudeDir); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf(".claude/ is a symlink, refusing to remove for safety")
		}
	}

	// Handle removal based on mode
	if mode == ModeUpdate {
		if err := selectiveRemove(claudeDir); err != nil {
			return nil, fmt.Errorf("selective remove: %w", err)
		}
	} else if force {
		_ = os.RemoveAll(claudeDir)
	}

	// Extract ZIP into target directory
	stats, err := unzip(result.Bundle, targetDir)
	if err != nil {
		return nil, fmt.Errorf("extract bundle: %w", err)
	}

	// Write .claudex.lock
	lock := &LockData{
		Version:     result.Version,
		Plan:        plan,
		InstalledAt: time.Now().UTC(),
		Checksum:    result.Checksum,
		CLIVersion:  cliVersion,
	}
	if err := WriteLock(targetDir, lock); err != nil {
		return nil, err
	}

	return stats, nil
}

// selectiveRemove backs up preserved paths, removes .claude/, then restores them.
func selectiveRemove(claudeDir string) (retErr error) {
	tmpDir, err := os.MkdirTemp("", "claudex-preserve-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer func() {
		if retErr == nil {
			os.RemoveAll(tmpDir)
		} else {
			fmt.Fprintf(os.Stderr, "  WARNING: preserved files backed up at %s\n", tmpDir)
		}
	}()

	// Backup preserved paths
	for _, rel := range preservePaths {
		src := filepath.Join(claudeDir, rel)
		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue
		}
		dst := filepath.Join(tmpDir, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return err
		}
		if err := copyPath(src, dst); err != nil {
			return fmt.Errorf("backup %s: %w", rel, err)
		}
	}

	// Remove .claude/
	if err := os.RemoveAll(claudeDir); err != nil {
		return fmt.Errorf("remove .claude/: %w", err)
	}

	// Restore preserved paths
	for _, rel := range preservePaths {
		src := filepath.Join(tmpDir, rel)
		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue
		}
		dst := filepath.Join(claudeDir, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return err
		}
		if err := copyPath(src, dst); err != nil {
			return fmt.Errorf("restore %s: %w", rel, err)
		}
	}

	return nil
}

// copyPath copies a file or directory recursively.
func copyPath(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0600)
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if err := copyPath(srcPath, dstPath); err != nil {
			return err
		}
	}
	return nil
}

// unzip extracts a ZIP byte slice into targetDir and counts installed items.
func unzip(zipBytes []byte, targetDir string) (*FileStats, error) {
	absDir, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, fmt.Errorf("resolve target dir: %w", err)
	}

	r, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return nil, fmt.Errorf("zip reader: %w", err)
	}

	stats := &FileStats{}

	for _, f := range r.File {
		// Path traversal protection: reject absolute paths and ..
		cleanName := filepath.Clean(f.Name)
		if filepath.IsAbs(cleanName) {
			continue
		}

		target := filepath.Join(targetDir, cleanName)
		absTarget, err := filepath.Abs(target)
		if err != nil {
			continue
		}
		if !strings.HasPrefix(absTarget, absDir+string(filepath.Separator)) && absTarget != absDir {
			continue
		}

		// Reject symlinks
		if f.FileInfo().Mode()&os.ModeSymlink != 0 {
			continue
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return nil, fmt.Errorf("mkdir %s: %w", target, err)
			}
			continue
		}

		// Skip root-level files that belong to the project (e.g. .gitignore)
		if shouldSkipOverwrite(cleanName, targetDir) {
			continue
		}

		// Create parent dirs
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return nil, fmt.Errorf("mkdir parent %s: %w", target, err)
		}

		// Extract file with size limit — sanitize permissions (no SUID/SGID/execute)
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", f.Name, err)
		}

		perm := f.Mode().Perm() & 0666
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
		if err != nil {
			rc.Close()
			return nil, fmt.Errorf("create %s: %w", target, err)
		}

		_, err = io.Copy(out, io.LimitReader(rc, maxFileSize))
		out.Close()
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("write %s: %w", target, err)
		}

		trackFile(cleanName, stats)
	}

	return stats, nil
}

// shouldSkipOverwrite returns true if the file is in neverOverwrite list and already exists.
func shouldSkipOverwrite(cleanName, targetDir string) bool {
	for _, name := range neverOverwrite {
		if cleanName == name {
			target := filepath.Join(targetDir, cleanName)
			if _, err := os.Stat(target); err == nil {
				return true
			}
		}
	}
	return false
}

// trackFile increments stats counters based on file path.
// Counts only top-level entries:
//   - Skills: .claude/skills/<slug>/SKILL.md
//   - Agents: .claude/agents/<slug>.md (single file, no nesting)
//   - Rules:  .claude/rules/<slug>.md (excludes per-skill nested rules/)
func trackFile(name string, stats *FileStats) {
	slug := filepath.ToSlash(name)

	if slug == "CLAUDE.md" || strings.HasSuffix(slug, "/CLAUDE.md") {
		stats.HasClaudeMD = true
		return
	}

	switch {
	case strings.HasPrefix(slug, ".claude/skills/") && strings.HasSuffix(slug, "/SKILL.md"):
		stats.SkillCount++
	case strings.HasPrefix(slug, ".claude/agents/") && strings.HasSuffix(slug, ".md"):
		if rest := strings.TrimPrefix(slug, ".claude/agents/"); !strings.Contains(rest, "/") {
			stats.AgentCount++
		}
	case strings.HasPrefix(slug, ".claude/rules/") && strings.HasSuffix(slug, ".md"):
		if rest := strings.TrimPrefix(slug, ".claude/rules/"); !strings.Contains(rest, "/") {
			stats.RuleCount++
		}
	}
}
