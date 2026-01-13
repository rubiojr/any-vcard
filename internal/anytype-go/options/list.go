// filepath: /home/epheo/dev/anytype-go/pkg/anytype/options/list.go
package options

// ListOptions contains the options for listing objects
type ListOptions struct {
	Limit  int
	Offset int
}

// ListOption is a function that configures ListOptions
type ListOption func(*ListOptions)

// WithLimit sets the maximum number of items to return
func WithLimit(limit int) ListOption {
	return func(opts *ListOptions) {
		opts.Limit = limit
	}
}

// WithOffset sets the number of items to skip before starting to collect the result set
func WithOffset(offset int) ListOption {
	return func(opts *ListOptions) {
		opts.Offset = offset
	}
}

// ApplyListOptions applies the given ListOptions to the ListOptions struct
func ApplyListOptions(opts ...ListOption) ListOptions {
	var options ListOptions
	for _, opt := range opts {
		opt(&options)
	}
	return options
}
