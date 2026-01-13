package anytype

import (
	"context"
)

// TemplateClient provides operations on templates for a specific type
type TemplateClient interface {
	// List retrieves all templates for a type
	List(ctx context.Context) ([]Template, error)

	// Get retrieves a specific template by ID
	Get(ctx context.Context, templateID string) (*Template, error)
}

// Template represents a template for creating objects
type Template struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Icon     *Icon  `json:"icon,omitempty"`
	Archived bool   `json:"archived"`
}

// TemplateContext provides operations on a specific template
type TemplateContext interface {
	// Get retrieves details of this specific template
	Get(ctx context.Context) (*TemplateResponse, error)
}

// TemplateResponse represents the response from a Get call on a template
type TemplateResponse struct {
	Template Template `json:"template"`
}
