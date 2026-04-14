// Package projects manages the central registry of projects that ran claudex init.
package projects

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Project represents a tracked project entry.
type Project struct {
	Path        string    `json:"path"`
	Version     string    `json:"version"`
	InstalledAt time.Time `json:"installed_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Store holds the list of tracked projects and the file path for persistence.
type Store struct {
	Projects []Project `json:"projects"`
	filePath string
}

// NewStore creates an empty store with the given file path.
func NewStore(filePath string) *Store {
	return &Store{filePath: filePath}
}

// Load reads projects.json from the given path. Returns empty store if file doesn't exist.
func Load(filePath string) (*Store, error) {
	s := &Store{filePath: filePath}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, fmt.Errorf("read projects file: %w", err)
	}

	if err := json.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("parse projects file: %w", err)
	}
	return s, nil
}

// Save writes the store to disk with 0600 permissions.
func (s *Store) Save() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal projects: %w", err)
	}
	return os.WriteFile(s.filePath, data, 0600)
}

// Register adds or updates a project entry. Uses absolute path for consistency.
func (s *Store) Register(projectPath, version string) error {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}
	absPath = filepath.Clean(absPath)
	now := time.Now().UTC()

	for i, p := range s.Projects {
		if filepath.Clean(p.Path) == absPath {
			s.Projects[i].Version = version
			s.Projects[i].UpdatedAt = now
			return nil
		}
	}

	s.Projects = append(s.Projects, Project{
		Path:        absPath,
		Version:     version,
		InstalledAt: now,
		UpdatedAt:   now,
	})
	return nil
}

// UpdateVersion updates version and timestamp for an existing project.
// Delegates to Register which handles both insert and update.
func (s *Store) UpdateVersion(projectPath, version string) error {
	return s.Register(projectPath, version)
}

// CleanStale removes entries where the project path no longer exists.
// Returns the number of removed entries.
func (s *Store) CleanStale() int {
	cleaned := 0
	valid := make([]Project, 0, len(s.Projects))

	for _, p := range s.Projects {
		if _, err := os.Stat(p.Path); err == nil {
			valid = append(valid, p)
		} else if os.IsNotExist(err) {
			cleaned++
		} else {
			// Permission error or other — keep the entry
			valid = append(valid, p)
		}
	}
	s.Projects = valid
	return cleaned
}

// List returns all tracked projects.
func (s *Store) List() []Project {
	return s.Projects
}

// ValidProjects returns projects where path still exists and has .claude/ dir.
func (s *Store) ValidProjects() []Project {
	valid := make([]Project, 0, len(s.Projects))
	for _, p := range s.Projects {
		claudeDir := filepath.Join(p.Path, ".claude")
		if _, err := os.Stat(claudeDir); err == nil {
			valid = append(valid, p)
		}
	}
	return valid
}
