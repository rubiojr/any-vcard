package mocks

import (
	"context"

	"github.com/epheo/anytype-go"
)

// MockMembersService implements the anytype.MemberClient interface for testing
type MockMembersService struct {
	ListFunc func(ctx context.Context) (*anytype.MemberListResponse, error)
}

// NewMockMembersService creates a new instance of MockMembersService with default implementations
func NewMockMembersService() *MockMembersService {
	return &MockMembersService{
		ListFunc: func(ctx context.Context) (*anytype.MemberListResponse, error) {
			return &anytype.MemberListResponse{
				Data: []anytype.Member{
					{
						ID:     "mock-member-id",
						Name:   "Mock User",
						Role:   "owner",
						Status: "active",
					},
				},
			}, nil
		},
	}
}

// List calls the mock implementation
func (s *MockMembersService) List(ctx context.Context) (*anytype.MemberListResponse, error) {
	return s.ListFunc(ctx)
}

// Member returns a mock member service
func (s *MockMembersService) Member(memberID string) anytype.MemberContext {
	return NewMockMemberService(memberID)
}

// MockMemberService implements the anytype.MemberContext interface for testing
type MockMemberService struct {
	MemberID string
	GetFunc  func(ctx context.Context) (*anytype.MemberResponse, error)
}

// NewMockMemberService creates a new instance of MockMemberService with default implementations
func NewMockMemberService(memberID string) *MockMemberService {
	return &MockMemberService{
		MemberID: memberID,
		GetFunc: func(ctx context.Context) (*anytype.MemberResponse, error) {
			return &anytype.MemberResponse{
				Member: anytype.Member{
					ID:     memberID,
					Name:   "Mock User",
					Role:   "owner",
					Status: "active",
				},
			}, nil
		},
	}
}

// Get calls the mock implementation
func (s *MockMemberService) Get(ctx context.Context) (*anytype.MemberResponse, error) {
	return s.GetFunc(ctx)
}

// MockListService implements the anytype.ListContext interface for testing
type MockListService struct {
	ListID      string
	ViewsFunc   func() anytype.ViewClient
	ObjectsFunc func() anytype.ObjectListClient
	ViewFunc    func(viewID string) anytype.ViewContext
	ObjectFunc  func(objectID string) anytype.ObjectListContext
}

// NewMockListService creates a new instance of MockListService with default implementations
func NewMockListService(listID string) *MockListService {
	return &MockListService{
		ListID: listID,
		ViewsFunc: func() anytype.ViewClient {
			return NewMockViewsService()
		},
		ObjectsFunc: func() anytype.ObjectListClient {
			return NewMockListObjectsService()
		},
		ViewFunc: func(viewID string) anytype.ViewContext {
			return NewMockViewService(viewID)
		},
		ObjectFunc: func(objectID string) anytype.ObjectListContext {
			return NewMockListObjectService(objectID)
		},
	}
}

// Views returns a mock views service
func (s *MockListService) Views() anytype.ViewClient {
	return s.ViewsFunc()
}

// Objects returns a mock list objects service
func (s *MockListService) Objects() anytype.ObjectListClient {
	return s.ObjectsFunc()
}

// View returns a mock view service
func (s *MockListService) View(viewID string) anytype.ViewContext {
	return s.ViewFunc(viewID)
}

// Object returns a mock list object service
func (s *MockListService) Object(objectID string) anytype.ObjectListContext {
	return s.ObjectFunc(objectID)
}

// MockViewsService implements the anytype.ViewClient interface for testing
type MockViewsService struct {
	ListFunc func(ctx context.Context) (*anytype.ViewListResponse, error)
}

// NewMockViewsService creates a new instance of MockViewsService with default implementations
func NewMockViewsService() *MockViewsService {
	return &MockViewsService{
		ListFunc: func(ctx context.Context) (*anytype.ViewListResponse, error) {
			return &anytype.ViewListResponse{
				Data: []anytype.ListView{
					{
						ID:   "mock-view-id",
						Name: "Default View",
					},
				},
			}, nil
		},
	}
}

// List calls the mock implementation
func (s *MockViewsService) List(ctx context.Context) (*anytype.ViewListResponse, error) {
	return s.ListFunc(ctx)
}

// MockViewService implements the anytype.ViewContext interface for testing
type MockViewService struct {
	ViewID         string
	ObjectsService *MockViewObjectsService
}

// NewMockViewService creates a new instance of MockViewService with default implementations
func NewMockViewService(viewID string) *MockViewService {
	return &MockViewService{
		ViewID:         viewID,
		ObjectsService: NewMockViewObjectsService(),
	}
}

// Objects returns a mock view objects service
func (s *MockViewService) Objects() anytype.ObjectViewClient {
	return s.ObjectsService
}

// MockViewObjectsService implements the anytype.ObjectViewClient interface for testing
type MockViewObjectsService struct {
	ListFunc func(ctx context.Context) (*anytype.ObjectListResponse, error)
}

// NewMockViewObjectsService creates a new instance of MockViewObjectsService with default implementations
func NewMockViewObjectsService() *MockViewObjectsService {
	return &MockViewObjectsService{
		ListFunc: func(ctx context.Context) (*anytype.ObjectListResponse, error) {
			return &anytype.ObjectListResponse{
				Data: []anytype.Object{
					{
						ID:      "mock-list-object-id-1",
						Name:    "List Item 1",
						TypeKey: "ot-page",
					},
					{
						ID:      "mock-list-object-id-2",
						Name:    "List Item 2",
						TypeKey: "ot-page",
					},
				},
			}, nil
		},
	}
}

// List calls the mock implementation
func (s *MockViewObjectsService) List(ctx context.Context) (*anytype.ObjectListResponse, error) {
	return s.ListFunc(ctx)
}

// MockListObjectsService implements the anytype.ObjectListClient interface for testing
type MockListObjectsService struct {
	AddFunc  func(ctx context.Context, objectIDs []string) error
	ListFunc func(ctx context.Context) (*anytype.ObjectListResponse, error)
}

// NewMockListObjectsService creates a new instance of MockListObjectsService with default implementations
func NewMockListObjectsService() *MockListObjectsService {
	return &MockListObjectsService{
		AddFunc: func(ctx context.Context, objectIDs []string) error {
			return nil
		},
		ListFunc: func(ctx context.Context) (*anytype.ObjectListResponse, error) {
			return &anytype.ObjectListResponse{
				Data: []anytype.Object{
					{
						ID:      "mock-list-object-id-1",
						Name:    "List Item 1",
						TypeKey: "ot-page",
					},
					{
						ID:      "mock-list-object-id-2",
						Name:    "List Item 2",
						TypeKey: "ot-page",
					},
				},
			}, nil
		},
	}
}

// Add calls the mock implementation
func (s *MockListObjectsService) Add(ctx context.Context, objectIDs []string) error {
	return s.AddFunc(ctx, objectIDs)
}

// List calls the mock implementation
func (s *MockListObjectsService) List(ctx context.Context) (*anytype.ObjectListResponse, error) {
	return s.ListFunc(ctx)
}

// MockListObjectService implements the anytype.ObjectListContext interface for testing
type MockListObjectService struct {
	ObjectID   string
	RemoveFunc func(ctx context.Context) error
}

// NewMockListObjectService creates a new instance of MockListObjectService with default implementations
func NewMockListObjectService(objectID string) *MockListObjectService {
	return &MockListObjectService{
		ObjectID: objectID,
		RemoveFunc: func(ctx context.Context) error {
			return nil
		},
	}
}

// Remove calls the mock implementation
func (s *MockListObjectService) Remove(ctx context.Context) error {
	return s.RemoveFunc(ctx)
}
