package anytype

import (
	"context"
)

// BlockClient provides operations on blocks within an object
type BlockClient interface {
	// List returns all blocks in the object
	List(ctx context.Context) ([]Block, error)

	// Get retrieves a specific block by ID
	Get(ctx context.Context, blockID string) (*Block, error)

	// Create creates a new block in the object
	Create(ctx context.Context, request CreateBlockRequest) (*Block, error)

	// Update updates a block in the object
	Update(ctx context.Context, blockID string, request UpdateBlockRequest) error

	// Delete deletes a block from the object
	Delete(ctx context.Context, blockID string) error
}

// Block represents a content block within an object
type Block struct {
	ID              string
	Text            *Text
	File            *File
	Property        *Property
	ChildrenIDs     []string `json:"children_ids"`
	Align           string
	VerticalAlign   string `json:"vertical_align"`
	BackgroundColor string `json:"background_color"`
}

// Text represents text content within a block
type Text struct {
	Content  string
	Style    string
	Markdown string
}

// File represents a file reference
type File struct {
	Name           string
	Hash           string
	Mime           string
	Size           int64
	Type           string
	State          string
	Style          string
	AddedAt        int64  `json:"added_at"`
	TargetObjectID string `json:"target_object_id"`
}

// CreateBlockRequest contains parameters for creating a new block
type CreateBlockRequest struct {
	Text           *Text
	File           *File
	Property       *Property
	ParentID       string `json:"parent_id"`
	TargetPosition int    `json:"target_position"`
}

// UpdateBlockRequest contains parameters for updating a block
type UpdateBlockRequest struct {
	Text            *Text
	File            *File
	Property        *Property
	Align           string
	VerticalAlign   string `json:"vertical_align"`
	BackgroundColor string `json:"background_color"`
}
