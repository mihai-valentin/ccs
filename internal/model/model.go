package model

import "time"

// Session represents an indexed Claude Code session.
type Session struct {
	ID           string
	ProjectDir   string
	Cwd          string
	GitBranch    string // may be empty if NULL in DB
	Name         string // may be empty if NULL in DB
	FirstMessage string // may be empty if NULL in DB
	LastMessage  string // may be empty if NULL in DB
	MessageCount int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	FileSize     int64
	FileModTime  time.Time
	Summary      string
	Tags         []Tag
}

// Tag represents a user-defined label for sessions.
type Tag struct {
	ID   int64
	Name string
}

// SessionFilter holds optional filters for listing sessions.
type SessionFilter struct {
	ProjectDir string
	Tags       []string
	Limit      int
	SortBy     string // "updated", "created", "name"
}
