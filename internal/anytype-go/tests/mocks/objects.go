package mocks

import (
	"context"

	"github.com/epheo/anytype-go"
	"github.com/epheo/anytype-go/options"
)

// MockObjectsService implements the anytype.ObjectClient interface for testing
type MockObjectsService struct {
	CurrentObjectID string
	ListFunc        func(ctx context.Context, opts ...options.ListOption) ([]anytype.Object, error)
	CreateFunc      func(ctx context.Context, req anytype.CreateObjectRequest) (*anytype.ObjectResponse, error)
	GetFunc         func(ctx context.Context) (*anytype.ObjectResponse, error)
	ExportFunc      func(ctx context.Context, format string) (*anytype.ExportResult, error)
	DeleteFunc      func(ctx context.Context) (*anytype.ObjectResponse, error)
	UpdateFunc      func(ctx context.Context, req anytype.UpdateObjectRequest) (*anytype.ObjectResponse, error)
}

// NewMockObjectsService creates a new instance of MockObjectsService with default implementations
func NewMockObjectsService() *MockObjectsService {
	return &MockObjectsService{
		ListFunc: func(ctx context.Context, opts ...options.ListOption) ([]anytype.Object, error) {
			return []anytype.Object{
				{
					ID:      "mock-object-id-1",
					Name:    "Mock Object 1",
					SpaceID: "mock-space-id",
					TypeKey: "page",
					Layout:  "basic",
				},
				{
					ID:      "mock-object-id-2",
					Name:    "Mock Object 2",
					SpaceID: "mock-space-id",
					TypeKey: "page",
					Layout:  "basic",
				},
			}, nil
		},
		CreateFunc: func(ctx context.Context, req anytype.CreateObjectRequest) (*anytype.ObjectResponse, error) {
			// Use the provided name or a default
			name := "New Mock Object"
			if req.Name != "" {
				name = req.Name
			}

			// Use the provided type or a default
			typeKey := "page"
			if req.TypeKey != "" {
				typeKey = req.TypeKey
			}

			var icon *anytype.Icon
			if req.Icon != nil {
				icon = req.Icon
			}

			return &anytype.ObjectResponse{
				Object: &anytype.Object{
					ID:      "new-mock-object-id",
					Name:    name,
					SpaceID: "mock-space-id",
					TypeKey: typeKey,
					Layout:  "basic",
					Icon:    icon,
				},
			}, nil
		},
		GetFunc: func(ctx context.Context) (*anytype.ObjectResponse, error) {
			// Default object ID
			objectID := "mock-object-id-1"

			return &anytype.ObjectResponse{
				Object: &anytype.Object{
					ID:      objectID,
					Name:    "Mock Object",
					SpaceID: "mock-space-id",
					TypeKey: "page",
					Layout:  "basic",
				},
			}, nil
		},
		ExportFunc: func(ctx context.Context, format string) (*anytype.ExportResult, error) {
			return &anytype.ExportResult{
				Markdown: "# Mock Object\n\nThis is the content of a mock object exported as markdown.",
			}, nil
		},
		DeleteFunc: func(ctx context.Context) (*anytype.ObjectResponse, error) {
			return &anytype.ObjectResponse{
				Object: &anytype.Object{
					ID:       "mock-object-id-1",
					Archived: true,
				},
			}, nil
		},
		UpdateFunc: func(ctx context.Context, req anytype.UpdateObjectRequest) (*anytype.ObjectResponse, error) {
			return &anytype.ObjectResponse{
				Object: &anytype.Object{
					ID:      "mock-object-id-1",
					Name:    "Updated Mock Object",
					SpaceID: "mock-space-id",
				},
			}, nil
		},
	}
}

// SetCurrentObjectID sets the current object ID for this service
func (s *MockObjectsService) SetCurrentObjectID(objectID string) {
	s.CurrentObjectID = objectID
}

// List calls the mock implementation
func (s *MockObjectsService) List(ctx context.Context, opts ...options.ListOption) ([]anytype.Object, error) {
	return s.ListFunc(ctx, opts...)
}

// Create calls the mock implementation
func (s *MockObjectsService) Create(ctx context.Context, req anytype.CreateObjectRequest) (*anytype.ObjectResponse, error) {
	return s.CreateFunc(ctx, req)
}

// Get calls the mock implementation
func (s *MockObjectsService) Get(ctx context.Context) (*anytype.ObjectResponse, error) {
	// Override the mock implementation to respect CurrentObjectID if it's set
	if s.CurrentObjectID != "" {
		return &anytype.ObjectResponse{
			Object: &anytype.Object{
				ID:      s.CurrentObjectID,
				Name:    "Mock Object",
				SpaceID: "mock-space-id",
				TypeKey: "page",
				Layout:  "basic",
			},
		}, nil
	}
	// Otherwise, use the default mock implementation
	return s.GetFunc(ctx)
}

// Export implements the ExportFunc for the mock
func (s *MockObjectsService) Export(ctx context.Context, format string) (*anytype.ExportResult, error) {
	if s.ExportFunc != nil {
		return s.ExportFunc(ctx, format)
	}

	// By default, return a mock markdown result
	return &anytype.ExportResult{
		Markdown: "# Mock Object\n\nThis is the content of a mock object exported as markdown.",
	}, nil
}

// Delete calls the mock implementation
func (s *MockObjectsService) Delete(ctx context.Context) (*anytype.ObjectResponse, error) {
	return s.DeleteFunc(ctx)
}

// Update calls the mock implementation
func (s *MockObjectsService) Update(ctx context.Context, req anytype.UpdateObjectRequest) (*anytype.ObjectResponse, error) {
	return s.UpdateFunc(ctx, req)
}

// Object returns a mock object service for a specific object
func (s *MockObjectsService) Object(objectID string) anytype.ObjectContext {
	s.SetCurrentObjectID(objectID)
	return s
}
