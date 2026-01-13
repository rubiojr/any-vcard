package anytype

import (
	"context"
)

// ListClient provides operations on lists within a space
type ListClient interface {
	// Add is used for adding objects to any list
	Add(ctx context.Context, objectIDs []string) error
}

// ListContext provides operations on a specific list
type ListContext interface {
	// Views returns a ViewClient for this list
	Views() ViewClient

	// View returns a ViewContext for a specific view in this list
	View(viewID string) ViewContext

	// Objects returns an ObjectListClient for this list
	Objects() ObjectListClient

	// Object returns an ObjectListContext for a specific object in this list
	Object(objectID string) ObjectListContext
}

// ViewClient provides operations on views within a list
type ViewClient interface {
	// List retrieves all views for the list
	List(ctx context.Context) (*ViewListResponse, error)
}

// ViewContext provides operations on a specific view
type ViewContext interface {
	// Objects returns an ObjectViewClient for this view
	Objects() ObjectViewClient
}

// ObjectListClient provides operations on objects within a list
type ObjectListClient interface {
	// List returns all objects in the list
	List(ctx context.Context) (*ObjectListResponse, error)

	// Add adds objects to the list
	Add(ctx context.Context, objectIDs []string) error
}

// ObjectListContext provides operations on a specific object within a list
type ObjectListContext interface {
	// Remove removes the object from the list
	Remove(ctx context.Context) error
}

// ObjectViewClient provides operations on objects within a view
type ObjectViewClient interface {
	// List returns all objects in the view
	List(ctx context.Context) (*ObjectListResponse, error)
}

// ListView represents a view configuration for a list
type ListView struct {
	ID      string       `json:"id"`
	Name    string       `json:"name"`
	Layout  string       `json:"layout"`
	Filters []ListFilter `json:"filters,omitempty"`
	Sorts   []ListSort   `json:"sorts,omitempty"`
}

// ListFilter represents a filter applied to a list view
type ListFilter struct {
	ID          string `json:"id"`
	PropertyKey string `json:"property_key"`
	Format      string `json:"format"`
	Condition   string `json:"condition"`
	Value       string `json:"value"`
}

// ListSort represents a sort applied to a list view
type ListSort struct {
	ID          string `json:"id"`
	PropertyKey string `json:"property_key"`
	Format      string `json:"format"`
	SortType    string `json:"sort_type"`
}

// ViewListResponse represents the paginated response for list views
type ViewListResponse struct {
	Data       []ListView `json:"data"`
	Pagination struct {
		Limit   int  `json:"limit"`
		Offset  int  `json:"offset"`
		Total   int  `json:"total"`
		HasMore bool `json:"has_more"`
	} `json:"pagination"`
}

// ObjectListResponse represents the paginated response for objects in a list
type ObjectListResponse struct {
	Data       []Object `json:"data"`
	Pagination struct {
		Limit   int  `json:"limit"`
		Offset  int  `json:"offset"`
		Total   int  `json:"total"`
		HasMore bool `json:"has_more"`
	} `json:"pagination"`
}
