package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/epheo/anytype-go"
)

var ErrTypeNotFound = errors.New("type not found")

// TypeClientImpl implements the TypeClient interface
type TypeClientImpl struct {
	client  *ClientImpl
	spaceID string
}

// List returns all available object types in the space
func (tc *TypeClientImpl) List(ctx context.Context) ([]anytype.Type, error) {
	// Make an HTTP request to GET /spaces/{space_id}/types
	req, err := tc.client.newRequest(ctx, "GET", fmt.Sprintf("/spaces/%s/types", tc.spaceID), nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data []anytype.Type `json:"data"`
	}
	if err := tc.client.doRequest(req, &response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

// Get retrieves details of a specific type by key
func (tc *TypeClientImpl) Get(ctx context.Context, typeKey string) (*anytype.Type, error) {
	// Make an HTTP request to GET /spaces/{space_id}/types/{type_key}
	req, err := tc.client.newRequest(ctx, "GET", fmt.Sprintf("/spaces/%s/types/%s", tc.spaceID, typeKey), nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Type anytype.Type `json:"type"`
	}
	if err := tc.client.doRequest(req, &response); err != nil {
		return nil, err
	}

	return &response.Type, nil
}

// GetKeyByName looks up a type key by its name
func (tc *TypeClientImpl) GetKeyByName(ctx context.Context, name string) (string, error) {
	// Actual implementation would list all types and find the one matching the name
	types, err := tc.List(ctx)
	if err != nil {
		return "", err
	}

	for _, t := range types {
		if t.Name == name {
			return t.Key, nil
		}
	}

	return "", ErrTypeNotFound
}

// Create creates a new type in the space
func (tc *TypeClientImpl) Create(ctx context.Context, request anytype.CreateTypeRequest) (*anytype.TypeResponse, error) {
	endpoint := fmt.Sprintf("/spaces/%s/types", tc.spaceID)

	req, err := tc.client.newRequest(ctx, http.MethodPost, endpoint, request)
	if err != nil {
		return nil, err
	}

	var response anytype.TypeResponse
	err = tc.client.doRequest(req, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// Type returns a TypeContext for a specific type
func (tc *TypeClientImpl) Type(typeID string) anytype.TypeContext {
	return &TypeContextImpl{
		client:  tc.client,
		spaceID: tc.spaceID,
		typeID:  typeID,
	}
}

// TypeContextImpl implements the TypeContext interface
type TypeContextImpl struct {
	client  *ClientImpl
	spaceID string
	typeID  string
}

// Get retrieves details of this specific type
func (tc *TypeContextImpl) Get(ctx context.Context) (*anytype.TypeResponse, error) {
	// Make an HTTP request to GET /spaces/{space_id}/types/{type_id}
	req, err := tc.client.newRequest(ctx, "GET", fmt.Sprintf("/spaces/%s/types/%s", tc.spaceID, tc.typeID), nil)
	if err != nil {
		return nil, err
	}

	var response anytype.TypeResponse
	if err := tc.client.doRequest(req, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// Templates returns a TemplateClient for this type
func (tc *TypeContextImpl) Templates() anytype.TemplateClient {
	return &TemplateClientImpl{
		client:  tc.client,
		spaceID: tc.spaceID,
		typeID:  tc.typeID,
	}
}

// Template returns a TemplateContext for a specific template of this type
func (tc *TypeContextImpl) Template(templateID string) anytype.TemplateContext {
	return &TemplateContextImpl{
		client:     tc.client,
		spaceID:    tc.spaceID,
		typeID:     tc.typeID,
		templateID: templateID,
	}
}
