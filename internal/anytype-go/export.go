package anytype

// ExportResult represents the result of an object export operation
type ExportResult struct {
	Markdown string `json:"markdown,omitempty"`
	// Add other export format fields as they become available
}
