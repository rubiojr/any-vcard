package anytype

// ClientOptions contains configuration options for the Anytype client
type ClientOptions struct {
	BaseURL string
	AppKey  string
}

// Client is the main interface for interacting with the Anytype API
type Client interface {
	// Auth returns an AuthClient for authentication operations
	Auth() AuthClient

	// Spaces returns a SpaceClient for working with spaces
	Spaces() SpaceClient

	// Space returns a specific SpaceContext for working with a given space
	Space(spaceID string) SpaceContext

	// Search returns a SearchClient for global search operations
	Search() SearchClient
}

// clientConstructor is a function type that constructs a Client
type clientConstructor func(ClientOptions) Client

// defaultClientConstructor is the default constructor for Client
var defaultClientConstructor clientConstructor

// RegisterClientConstructor registers a constructor function for creating Client instances
func RegisterClientConstructor(constructor clientConstructor) {
	defaultClientConstructor = constructor
}

// ClientOption is a function type that modifies ClientOptions
type ClientOption func(*ClientOptions)

// WithBaseURL sets the base URL for API requests
func WithBaseURL(url string) ClientOption {
	return func(o *ClientOptions) {
		o.BaseURL = url
	}
}

// WithAppKey sets the app key for authentication
func WithAppKey(appKey string) ClientOption {
	return func(o *ClientOptions) {
		o.AppKey = appKey
	}
}

// NewClient creates a new Anytype API client with the given options
func NewClient(opts ...ClientOption) Client {
	if defaultClientConstructor == nil {
		panic("No client constructor registered. Import the client implementation package.")
	}

	// Initialize with default options
	clientOpts := ClientOptions{}

	// Apply all provided options
	for _, opt := range opts {
		opt(&clientOpts)
	}

	client := defaultClientConstructor(clientOpts)
	return client
}

// These interfaces and structs have been moved to space.go for better organization
