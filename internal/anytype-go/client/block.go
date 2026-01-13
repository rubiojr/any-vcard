package client

import (
	"context"

	"github.com/epheo/anytype-go"
)

// BlockClientImpl implements the BlockClient interface
type BlockClientImpl struct {
	client   *ClientImpl
	spaceID  string
	objectID string
}

// List returns all blocks in the object
func (bc *BlockClientImpl) List(ctx context.Context) ([]anytype.Block, error) {
	// Actual implementation would make an HTTP request to the endpoint
	// GET /spaces/{space_id}/objects/{object_id}/blocks
	return nil, nil
}

// Get retrieves a specific block by ID
func (bc *BlockClientImpl) Get(ctx context.Context, blockID string) (*anytype.Block, error) {
	// Actual implementation would make an HTTP request to the endpoint
	// GET /spaces/{space_id}/objects/{object_id}/blocks/{block_id}
	return nil, nil
}

// Create creates a new block in the object
func (bc *BlockClientImpl) Create(ctx context.Context, request anytype.CreateBlockRequest) (*anytype.Block, error) {
	// Actual implementation would make an HTTP request to the endpoint
	// POST /spaces/{space_id}/objects/{object_id}/blocks
	return nil, nil
}

// Update updates a block in the object
func (bc *BlockClientImpl) Update(ctx context.Context, blockID string, request anytype.UpdateBlockRequest) error {
	// Actual implementation would make an HTTP request to the endpoint
	// PUT /spaces/{space_id}/objects/{object_id}/blocks/{block_id}
	return nil
}

// Delete deletes a block from the object
func (bc *BlockClientImpl) Delete(ctx context.Context, blockID string) error {
	// Actual implementation would make an HTTP request to the endpoint
	// DELETE /spaces/{space_id}/objects/{object_id}/blocks/{block_id}
	return nil
}
