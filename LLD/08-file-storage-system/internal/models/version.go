package models

import "time"

// Version represents a version of a file for version history.
type Version struct {
	ID            string
	FileID        string
	VersionNumber int
	Content       []byte
	Size          int64
	CreatedAt     time.Time
	CreatedBy     string
}
