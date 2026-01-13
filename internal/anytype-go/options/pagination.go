package options

// PaginationMetadata contains information about pagination
type PaginationMetadata struct {
	Total   int  `json:"total"`
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	HasMore bool `json:"has_more"`
}
