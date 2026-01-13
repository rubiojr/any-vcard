package client

import (
	"net/http"

	"github.com/epheo/anytype-go"
)

// ClientImpl is the actual implementation of the Client interface
type ClientImpl struct {
	httpClient *http.Client
	baseURL    string
	appKey     string
}

func init() {
	// Register our constructor with the main package
	anytype.RegisterClientConstructor(NewClient)
}

// NewClient creates a new Anytype API client with the given options
func NewClient(options anytype.ClientOptions) anytype.Client {
	return &ClientImpl{
		httpClient: http.DefaultClient,
		baseURL:    options.BaseURL,
		appKey:     options.AppKey,
	}
}

// Spaces returns a SpaceClient for working with spaces
func (c *ClientImpl) Spaces() anytype.SpaceClient {
	return &SpaceClientImpl{client: c}
}

// Space returns a SpaceContext for working with a specific space
func (c *ClientImpl) Space(spaceID string) anytype.SpaceContext {
	return &SpaceContextImpl{
		client:  c,
		spaceID: spaceID,
	}
}

// Search returns a SearchClient for global search operations
func (c *ClientImpl) Search() anytype.SearchClient {
	return &SearchClientImpl{client: c}
}

// Auth returns an AuthClient for authentication operations
func (c *ClientImpl) Auth() anytype.AuthClient {
	return &AuthClientImpl{client: c}
}
