package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Comment is one document comment row, addressable by its library position
// (LibraryCommentID) and optionally threaded under a parent library comment.
type Comment struct {
	ID               uuid.UUID
	TenantID         uuid.UUID
	DocumentID       uuid.UUID
	LibraryCommentID int
	ParentLibraryID  *int
	AuthorID         string
	AuthorDisplay    string
	ContentJSON      json.RawMessage
	ResolvedAt       *time.Time
	ResolvedBy       *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// CommentCreateInput carries the fields needed to create a new Comment.
type CommentCreateInput struct {
	LibraryCommentID int
	ParentLibraryID  *int
	AuthorDisplay    string
	ContentJSON      json.RawMessage
}

// CommentUpdateInput carries the fields to patch onto an existing Comment;
// nil fields are left unchanged.
type CommentUpdateInput struct {
	ContentJSON *json.RawMessage
	Done        *bool // true sets resolved_at=now()+resolved_by, false clears
}
