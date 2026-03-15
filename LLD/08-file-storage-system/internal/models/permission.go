package models

import "time"

// PermissionLevel defines the access level for shared files.
type PermissionLevel string

const (
	PermissionView PermissionLevel = "View"
	PermissionEdit PermissionLevel = "Edit"
	PermissionOwner PermissionLevel = "Owner"
)

// Permission represents sharing access for a file or folder.
type Permission struct {
	ID        string
	FileID    string // Can be file or folder ID
	UserID    string
	Level     PermissionLevel
	GrantedAt time.Time
	GrantedBy string
}

// CanView returns true if this permission allows viewing.
func (p *Permission) CanView() bool {
	return p.Level == PermissionView || p.Level == PermissionEdit || p.Level == PermissionOwner
}

// CanEdit returns true if this permission allows editing.
func (p *Permission) CanEdit() bool {
	return p.Level == PermissionEdit || p.Level == PermissionOwner
}
