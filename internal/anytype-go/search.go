package anytype

import (
	"context"
)

// SearchClient provides global search operations across all spaces
type SearchClient interface {
	// Search searches for objects across all spaces
	Search(ctx context.Context, request SearchRequest) (*SearchResponse, error)
}

// SearchResponse represents the response from a search operation
type SearchResponse struct {
	Data []Object `json:"data"`
}

// SearchRequest represents a search query
type SearchRequest struct {
	Query string       `json:"query"`
	Types []string     `json:"types,omitempty"` // Object type keys or IDs to filter by
	Sort  *SortOptions `json:"sort,omitempty"`  // Use pointer so it's omitted when nil
}

// SortOptions represents sorting options for search results
type SortOptions struct {
	Property  SortProperty  `json:"property"`
	Direction SortDirection `json:"direction"`
}

// SortProperty represents the property to sort search results by
type SortProperty string

// SortDirection represents the direction to sort search results
type SortDirection string

const (
	// Sort properties
	SortPropertyCreatedDate      SortProperty = "created_date"
	SortPropertyLastModifiedDate SortProperty = "last_modified_date"
	SortPropertyLastOpenedDate   SortProperty = "last_opened_date"
	SortPropertyName             SortProperty = "name"

	// Sort directions
	SortDirectionAsc  SortDirection = "asc"
	SortDirectionDesc SortDirection = "desc"
)
