package mocks

import (
	"context"
	"log"

	"github.com/epheo/anytype-go"
)

// MockAuthService implements the anytype.AuthClient interface for testing
type MockAuthService struct {
	CreateChallengeFunc func(ctx context.Context, appName string) (*anytype.CreateChallengeResponse, error)
	CreateApiKeyFunc    func(ctx context.Context, challengeID, code string) (*anytype.CreateApiKeyResponse, error)
	DisplayCodeFunc     func(ctx context.Context, appName string) (*anytype.DisplayCodeResponse, error)
	GetTokenFunc        func(ctx context.Context, challengeID, code string) (*anytype.TokenResponse, error)
}

// NewMockAuthService creates a new instance of MockAuthService with default implementations
func NewMockAuthService() *MockAuthService {
	return &MockAuthService{
		CreateChallengeFunc: func(ctx context.Context, appName string) (*anytype.CreateChallengeResponse, error) {
			return &anytype.CreateChallengeResponse{
				ChallengeID: "mock-challenge-id",
			}, nil
		},
		CreateApiKeyFunc: func(ctx context.Context, challengeID, code string) (*anytype.CreateApiKeyResponse, error) {
			return &anytype.CreateApiKeyResponse{
				ApiKey: "mock-api-key",
			}, nil
		},
		DisplayCodeFunc: func(ctx context.Context, appName string) (*anytype.DisplayCodeResponse, error) {
			log.Println("Warning: DisplayCode is deprecated, use CreateChallenge instead")
			return &anytype.DisplayCodeResponse{
				ChallengeID: "mock-challenge-id",
			}, nil
		},
		GetTokenFunc: func(ctx context.Context, challengeID, code string) (*anytype.TokenResponse, error) {
			log.Println("Warning: GetToken is deprecated, use CreateApiKey instead")
			return &anytype.TokenResponse{
				AppKey: "mock-app-key",
			}, nil
		},
	}
}

// CreateChallenge calls the mock implementation
func (s *MockAuthService) CreateChallenge(ctx context.Context, appName string) (*anytype.CreateChallengeResponse, error) {
	return s.CreateChallengeFunc(ctx, appName)
}

// CreateApiKey calls the mock implementation
func (s *MockAuthService) CreateApiKey(ctx context.Context, challengeID, code string) (*anytype.CreateApiKeyResponse, error) {
	return s.CreateApiKeyFunc(ctx, challengeID, code)
}

// DisplayCode calls the mock implementation
// Deprecated: Use CreateChallenge instead
func (s *MockAuthService) DisplayCode(ctx context.Context, appName string) (*anytype.DisplayCodeResponse, error) {
	return s.DisplayCodeFunc(ctx, appName)
}

// GetToken calls the mock implementation
// Deprecated: Use CreateApiKey instead
func (s *MockAuthService) GetToken(ctx context.Context, challengeID, code string) (*anytype.TokenResponse, error) {
	return s.GetTokenFunc(ctx, challengeID, code)
}
