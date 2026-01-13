package anytype

import (
	"context"
)

// AuthClient provides operations for authentication
type AuthClient interface {
	// CreateChallenge initiates a secure authentication flow
	CreateChallenge(ctx context.Context, appName string) (*CreateChallengeResponse, error)

	// CreateApiKey completes the authentication flow by providing a code
	CreateApiKey(ctx context.Context, challengeID string, code string) (*CreateApiKeyResponse, error)

	// !! Deprecated: Use CreateChallenge instead
	DisplayCode(ctx context.Context, appName string) (*DisplayCodeResponse, error)

	// !! Deprecated: Use CreateApiKey instead
	GetToken(ctx context.Context, challengeID string, code string) (*TokenResponse, error)
}

// CreateChallengeResponse represents the response from the create challenge endpoint
type CreateChallengeResponse struct {
	ChallengeID string `json:"challenge_id"`
}

// CreateApiKeyResponse represents the response from the create API key endpoint
type CreateApiKeyResponse struct {
	ApiKey string `json:"api_key"`
}

// DisplayCodeResponse represents the response from the display_code endpoint
// !! Deprecated: Use CreateChallengeResponse instead
type DisplayCodeResponse struct {
	ChallengeID string `json:"challenge_id"`
}

// TokenResponse represents the response from the token endpoint
// !! Deprecated: Use CreateApiKeyResponse instead
type TokenResponse struct {
	AppKey string `json:"app_key"`
}
