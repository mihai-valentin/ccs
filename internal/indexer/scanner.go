package indexer

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SessionFile represents a discovered JSONL session file on disk.
type SessionFile struct {
	Path        string
	ProjectDir  string
	SessionID   string
	FileSize    int64
	FileModTime time.Time
}

// ScanSessions walks claudeDir/projects/*/*.jsonl and returns all discovered
// session files, skipping any files inside subagents/ subdirectories.
// The returned bool (partial) is true if any walk errors were encountered,
// meaning some files may have been missed. Callers should avoid purging
// sessions when partial is true.
func ScanSessions(claudeDir string) ([]SessionFile, bool, error) {
	projectsDir := filepath.Join(claudeDir, "projects")

	info, err := os.Stat(projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if !info.IsDir() {
		return nil, false, nil
	}

	var files []SessionFile
	hadErrors := false

	err = filepath.WalkDir(projectsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Printf("warning: walk error at %s: %v", path, err)
			hadErrors = true
			return nil
		}

		// Skip subagents directories entirely.
		if d.IsDir() && d.Name() == "subagents" {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(d.Name(), ".jsonl") {
			return nil
		}

		// Session ID is the filename without extension.
		sessionID := strings.TrimSuffix(d.Name(), ".jsonl")

		// Project dir is the directory name containing this file,
		// relative to the projects/ directory.
		rel, err := filepath.Rel(projectsDir, filepath.Dir(path))
		if err != nil {
			return nil
		}

		fi, err := d.Info()
		if err != nil {
			return nil
		}

		files = append(files, SessionFile{
			Path:        path,
			ProjectDir:  rel,
			SessionID:   sessionID,
			FileSize:    fi.Size(),
			FileModTime: fi.ModTime(),
		})

		return nil
	})

	return files, hadErrors, err
}
