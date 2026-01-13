package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/epheo/anytype-go"
)

// TemplateClientImpl implements the TemplateClient interface
type TemplateClientImpl struct {
	client  *ClientImpl
	spaceID string
	typeID  string
}

// List retrieves all templates for a type
func (tc *TemplateClientImpl) List(ctx context.Context) ([]anytype.Template, error) {
	endpoint := fmt.Sprintf("/spaces/%s/types/%s/templates", tc.spaceID, tc.typeID)

	req, err := tc.client.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data       []anytype.Template `json:"data"`
		Pagination struct {
			HasMore bool `json:"has_more"`
			Limit   int  `json:"limit"`
			Offset  int  `json:"offset"`
			Total   int  `json:"total"`
		} `json:"pagination"`
	}

	if err := tc.client.doRequest(req, &response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

// Get retrieves a specific template by ID
func (tc *TemplateClientImpl) Get(ctx context.Context, templateID string) (*anytype.Template, error) {
	endpoint := fmt.Sprintf("/spaces/%s/types/%s/templates/%s", tc.spaceID, tc.typeID, templateID)

	req, err := tc.client.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Template anytype.Template `json:"template"`
	}

	if err := tc.client.doRequest(req, &response); err != nil {
		return nil, err
	}

	return &response.Template, nil
}

// TemplateContextImpl implements the TemplateContext interface
type TemplateContextImpl struct {
	client     *ClientImpl
	spaceID    string
	typeID     string
	templateID string
}

// Get retrieves details of this specific template
func (tc *TemplateContextImpl) Get(ctx context.Context) (*anytype.TemplateResponse, error) {
	endpoint := fmt.Sprintf("/spaces/%s/types/%s/templates/%s", tc.spaceID, tc.typeID, tc.templateID)

	req, err := tc.client.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Template anytype.Template `json:"template"`
	}

	if err := tc.client.doRequest(req, &response); err != nil {
		return nil, err
	}

	return &anytype.TemplateResponse{
		Template: response.Template,
	}, nil
}
