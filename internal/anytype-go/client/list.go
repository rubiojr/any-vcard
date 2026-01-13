package client

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/epheo/anytype-go"
)

// ListClientImpl implements the ListClient interface
type ListClientImpl struct {
	client  *ClientImpl
	spaceID string
	listID  string
}

// AddObjectsToListRequest represents the request structure for adding objects to a list
type AddObjectsToListRequest struct {
	Objects []string `json:"objects"`
}

// Add is used for adding objects to any list
func (lc *ListClientImpl) Add(ctx context.Context, objectIDs []string) error {
	// Create HTTP request
	endpoint := "/spaces/" + lc.spaceID + "/lists/" + lc.listID + "/objects"

	// Create the proper request structure
	requestBody := AddObjectsToListRequest{
		Objects: objectIDs,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	// Create HTTP request
	req, err := lc.client.newRequest(ctx, http.MethodPost, endpoint, jsonData)
	if err != nil {
		return err
	}

	// Make the request
	err = lc.client.doRequest(req, nil)
	if err != nil {
		return err
	}

	return nil
}

// GetViews retrieves views for a specific list
func (lc *ListClientImpl) GetViews(ctx context.Context, listID string) ([]anytype.ListView, error) {
	endpoint := "/spaces/" + lc.spaceID + "/lists/" + listID + "/views"
	req, err := lc.client.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	response := &anytype.ViewListResponse{}
	err = lc.client.doRequest(req, response)
	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

// GetObjects retrieves objects for a specific list view
func (lc *ListClientImpl) GetObjects(ctx context.Context, listID string, viewID string) ([]anytype.Object, error) {
	endpoint := "/spaces/" + lc.spaceID + "/lists/" + listID + "/views/" + viewID + "/objects"
	req, err := lc.client.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	response := &anytype.ObjectListResponse{}
	err = lc.client.doRequest(req, response)
	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

// AddObjects adds objects to a list
func (lc *ListClientImpl) AddObjects(ctx context.Context, listID string, objectIDs []string) error {
	endpoint := "/spaces/" + lc.spaceID + "/lists/" + listID + "/objects"

	// Create the proper request structure
	requestBody := AddObjectsToListRequest{
		Objects: objectIDs,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	req, err := lc.client.newRequest(ctx, http.MethodPost, endpoint, jsonData)
	if err != nil {
		return err
	}

	return lc.client.doRequest(req, nil)
}

// RemoveObject removes an object from a list
func (lc *ListClientImpl) RemoveObject(ctx context.Context, listID string, objectID string) error {
	endpoint := "/spaces/" + lc.spaceID + "/lists/" + listID + "/objects/" + objectID

	req, err := lc.client.newRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return lc.client.doRequest(req, nil)
}

// ListContextImpl implements the ListContext interface
type ListContextImpl struct {
	client  *ClientImpl
	spaceID string
	listID  string
}

// Views returns a ViewClient for this list
func (lc *ListContextImpl) Views() anytype.ViewClient {
	return &ViewClientImpl{
		client:  lc.client,
		spaceID: lc.spaceID,
		listID:  lc.listID,
	}
}

// View returns a ViewContext for a specific view in this list
func (lc *ListContextImpl) View(viewID string) anytype.ViewContext {
	return &ViewContextImpl{
		client:  lc.client,
		spaceID: lc.spaceID,
		listID:  lc.listID,
		viewID:  viewID,
	}
}

// Objects returns an ObjectListClient for this list
func (lc *ListContextImpl) Objects() anytype.ObjectListClient {
	return &ObjectListClientImpl{
		client:  lc.client,
		spaceID: lc.spaceID,
		listID:  lc.listID,
	}
}

// Object returns an ObjectListContext for a specific object in this list
func (lc *ListContextImpl) Object(objectID string) anytype.ObjectListContext {
	return &ObjectListContextImpl{
		client:   lc.client,
		spaceID:  lc.spaceID,
		listID:   lc.listID,
		objectID: objectID,
	}
}

// ViewClientImpl implements the ViewClient interface
type ViewClientImpl struct {
	client  *ClientImpl
	spaceID string
	listID  string
}

// List retrieves all views for the list
func (vc *ViewClientImpl) List(ctx context.Context) (*anytype.ViewListResponse, error) {
	endpoint := "/spaces/" + vc.spaceID + "/lists/" + vc.listID + "/views"
	req, err := vc.client.newRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	response := &anytype.ViewListResponse{}
	err = vc.client.doRequest(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// ViewContextImpl implements the ViewContext interface
type ViewContextImpl struct {
	client  *ClientImpl
	spaceID string
	listID  string
	viewID  string
}

// Objects returns an ObjectViewClient for this view
func (vc *ViewContextImpl) Objects() anytype.ObjectViewClient {
	return &ObjectViewClientImpl{
		client:  vc.client,
		spaceID: vc.spaceID,
		listID:  vc.listID,
		viewID:  vc.viewID,
	}
}

// ObjectListClientImpl implements the ObjectListClient interface
type ObjectListClientImpl struct {
	client  *ClientImpl
	spaceID string
	listID  string
}

// List returns all objects in the list
func (olc *ObjectListClientImpl) List(ctx context.Context) (*anytype.ObjectListResponse, error) {
	endpoint := "/spaces/" + olc.spaceID + "/lists/" + olc.listID + "/objects"
	req, err := olc.client.newRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	response := &anytype.ObjectListResponse{}
	err = olc.client.doRequest(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// Add adds objects to the list
func (olc *ObjectListClientImpl) Add(ctx context.Context, objectIDs []string) error {
	endpoint := "/spaces/" + olc.spaceID + "/lists/" + olc.listID + "/objects"

	// Create the proper request structure
	requestBody := AddObjectsToListRequest{
		Objects: objectIDs,
	}

	req, err := olc.client.newRequest(ctx, "POST", endpoint, requestBody)
	if err != nil {
		return err
	}

	return olc.client.doRequest(req, nil)
}

// ObjectListContextImpl implements the ObjectListContext interface
type ObjectListContextImpl struct {
	client   *ClientImpl
	spaceID  string
	listID   string
	objectID string
}

// Remove removes the object from the list
func (olc *ObjectListContextImpl) Remove(ctx context.Context) error {
	endpoint := "/spaces/" + olc.spaceID + "/lists/" + olc.listID + "/objects/" + olc.objectID
	req, err := olc.client.newRequest(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return err
	}

	return olc.client.doRequest(req, nil)
}

// ObjectViewClientImpl implements the ObjectViewClient interface
type ObjectViewClientImpl struct {
	client  *ClientImpl
	spaceID string
	listID  string
	viewID  string
}

// List returns all objects in the view
func (ovc *ObjectViewClientImpl) List(ctx context.Context) (*anytype.ObjectListResponse, error) {
	endpoint := "/spaces/" + ovc.spaceID + "/lists/" + ovc.listID + "/views/" + ovc.viewID + "/objects"
	req, err := ovc.client.newRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	response := &anytype.ObjectListResponse{}
	err = ovc.client.doRequest(req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
