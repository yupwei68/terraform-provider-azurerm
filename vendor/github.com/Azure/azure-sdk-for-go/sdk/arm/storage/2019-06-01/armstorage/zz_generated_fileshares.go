// +build go1.13

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

package armstorage

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/armcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"net/http"
	"net/url"
	"strings"
)

// FileSharesClient contains the methods for the FileShares group.
// Don't use this type directly, use NewFileSharesClient() instead.
type FileSharesClient struct {
	con            *armcore.Connection
	subscriptionID string
}

// NewFileSharesClient creates a new instance of FileSharesClient with the specified values.
func NewFileSharesClient(con *armcore.Connection, subscriptionID string) *FileSharesClient {
	return &FileSharesClient{con: con, subscriptionID: subscriptionID}
}

// Create - Creates a new share under the specified account as described by request body. The share resource includes metadata and properties for that share.
// It does not include a list of the files contained by
// the share.
func (client *FileSharesClient) Create(ctx context.Context, resourceGroupName string, accountName string, shareName string, fileShare FileShare, options *FileSharesCreateOptions) (FileShareResponse, error) {
	req, err := client.createCreateRequest(ctx, resourceGroupName, accountName, shareName, fileShare, options)
	if err != nil {
		return FileShareResponse{}, err
	}
	resp, err := client.con.Pipeline().Do(req)
	if err != nil {
		return FileShareResponse{}, err
	}
	if !resp.HasStatusCode(http.StatusOK, http.StatusCreated) {
		return FileShareResponse{}, client.createHandleError(resp)
	}
	return client.createHandleResponse(resp)
}

// createCreateRequest creates the Create request.
func (client *FileSharesClient) createCreateRequest(ctx context.Context, resourceGroupName string, accountName string, shareName string, fileShare FileShare, options *FileSharesCreateOptions) (*azcore.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Storage/storageAccounts/{accountName}/fileServices/default/shares/{shareName}"
	urlPath = strings.ReplaceAll(urlPath, "{resourceGroupName}", url.PathEscape(resourceGroupName))
	urlPath = strings.ReplaceAll(urlPath, "{accountName}", url.PathEscape(accountName))
	urlPath = strings.ReplaceAll(urlPath, "{shareName}", url.PathEscape(shareName))
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionId}", url.PathEscape(client.subscriptionID))
	req, err := azcore.NewRequest(ctx, http.MethodPut, azcore.JoinPaths(client.con.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	req.Telemetry(telemetryInfo)
	query := req.URL.Query()
	query.Set("api-version", "2019-06-01")
	req.URL.RawQuery = query.Encode()
	req.Header.Set("Accept", "application/json")
	return req, req.MarshalAsJSON(fileShare)
}

// createHandleResponse handles the Create response.
func (client *FileSharesClient) createHandleResponse(resp *azcore.Response) (FileShareResponse, error) {
	var val *FileShare
	if err := resp.UnmarshalAsJSON(&val); err != nil {
		return FileShareResponse{}, err
	}
	return FileShareResponse{RawResponse: resp.Response, FileShare: val}, nil
}

// createHandleError handles the Create error response.
func (client *FileSharesClient) createHandleError(resp *azcore.Response) error {
	var err CloudError
	if err := resp.UnmarshalAsJSON(&err); err != nil {
		return err
	}
	return azcore.NewResponseError(&err, resp.Response)
}

// Delete - Deletes specified share under its account.
func (client *FileSharesClient) Delete(ctx context.Context, resourceGroupName string, accountName string, shareName string, options *FileSharesDeleteOptions) (*http.Response, error) {
	req, err := client.deleteCreateRequest(ctx, resourceGroupName, accountName, shareName, options)
	if err != nil {
		return nil, err
	}
	resp, err := client.con.Pipeline().Do(req)
	if err != nil {
		return nil, err
	}
	if !resp.HasStatusCode(http.StatusOK, http.StatusNoContent) {
		return nil, client.deleteHandleError(resp)
	}
	return resp.Response, nil
}

// deleteCreateRequest creates the Delete request.
func (client *FileSharesClient) deleteCreateRequest(ctx context.Context, resourceGroupName string, accountName string, shareName string, options *FileSharesDeleteOptions) (*azcore.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Storage/storageAccounts/{accountName}/fileServices/default/shares/{shareName}"
	urlPath = strings.ReplaceAll(urlPath, "{resourceGroupName}", url.PathEscape(resourceGroupName))
	urlPath = strings.ReplaceAll(urlPath, "{accountName}", url.PathEscape(accountName))
	urlPath = strings.ReplaceAll(urlPath, "{shareName}", url.PathEscape(shareName))
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionId}", url.PathEscape(client.subscriptionID))
	req, err := azcore.NewRequest(ctx, http.MethodDelete, azcore.JoinPaths(client.con.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	req.Telemetry(telemetryInfo)
	query := req.URL.Query()
	query.Set("api-version", "2019-06-01")
	req.URL.RawQuery = query.Encode()
	req.Header.Set("Accept", "application/json")
	return req, nil
}

// deleteHandleError handles the Delete error response.
func (client *FileSharesClient) deleteHandleError(resp *azcore.Response) error {
	var err CloudError
	if err := resp.UnmarshalAsJSON(&err); err != nil {
		return err
	}
	return azcore.NewResponseError(&err, resp.Response)
}

// Get - Gets properties of a specified share.
func (client *FileSharesClient) Get(ctx context.Context, resourceGroupName string, accountName string, shareName string, options *FileSharesGetOptions) (FileShareResponse, error) {
	req, err := client.getCreateRequest(ctx, resourceGroupName, accountName, shareName, options)
	if err != nil {
		return FileShareResponse{}, err
	}
	resp, err := client.con.Pipeline().Do(req)
	if err != nil {
		return FileShareResponse{}, err
	}
	if !resp.HasStatusCode(http.StatusOK) {
		return FileShareResponse{}, client.getHandleError(resp)
	}
	return client.getHandleResponse(resp)
}

// getCreateRequest creates the Get request.
func (client *FileSharesClient) getCreateRequest(ctx context.Context, resourceGroupName string, accountName string, shareName string, options *FileSharesGetOptions) (*azcore.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Storage/storageAccounts/{accountName}/fileServices/default/shares/{shareName}"
	urlPath = strings.ReplaceAll(urlPath, "{resourceGroupName}", url.PathEscape(resourceGroupName))
	urlPath = strings.ReplaceAll(urlPath, "{accountName}", url.PathEscape(accountName))
	urlPath = strings.ReplaceAll(urlPath, "{shareName}", url.PathEscape(shareName))
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionId}", url.PathEscape(client.subscriptionID))
	req, err := azcore.NewRequest(ctx, http.MethodGet, azcore.JoinPaths(client.con.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	req.Telemetry(telemetryInfo)
	query := req.URL.Query()
	query.Set("api-version", "2019-06-01")
	if options != nil && options.Expand != nil {
		query.Set("$expand", "stats")
	}
	req.URL.RawQuery = query.Encode()
	req.Header.Set("Accept", "application/json")
	return req, nil
}

// getHandleResponse handles the Get response.
func (client *FileSharesClient) getHandleResponse(resp *azcore.Response) (FileShareResponse, error) {
	var val *FileShare
	if err := resp.UnmarshalAsJSON(&val); err != nil {
		return FileShareResponse{}, err
	}
	return FileShareResponse{RawResponse: resp.Response, FileShare: val}, nil
}

// getHandleError handles the Get error response.
func (client *FileSharesClient) getHandleError(resp *azcore.Response) error {
	var err CloudError
	if err := resp.UnmarshalAsJSON(&err); err != nil {
		return err
	}
	return azcore.NewResponseError(&err, resp.Response)
}

// List - Lists all shares.
func (client *FileSharesClient) List(resourceGroupName string, accountName string, options *FileSharesListOptions) FileShareItemsPager {
	return &fileShareItemsPager{
		pipeline: client.con.Pipeline(),
		requester: func(ctx context.Context) (*azcore.Request, error) {
			return client.listCreateRequest(ctx, resourceGroupName, accountName, options)
		},
		responder: client.listHandleResponse,
		errorer:   client.listHandleError,
		advancer: func(ctx context.Context, resp FileShareItemsResponse) (*azcore.Request, error) {
			return azcore.NewRequest(ctx, http.MethodGet, *resp.FileShareItems.NextLink)
		},
		statusCodes: []int{http.StatusOK},
	}
}

// listCreateRequest creates the List request.
func (client *FileSharesClient) listCreateRequest(ctx context.Context, resourceGroupName string, accountName string, options *FileSharesListOptions) (*azcore.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Storage/storageAccounts/{accountName}/fileServices/default/shares"
	urlPath = strings.ReplaceAll(urlPath, "{resourceGroupName}", url.PathEscape(resourceGroupName))
	urlPath = strings.ReplaceAll(urlPath, "{accountName}", url.PathEscape(accountName))
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionId}", url.PathEscape(client.subscriptionID))
	req, err := azcore.NewRequest(ctx, http.MethodGet, azcore.JoinPaths(client.con.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	req.Telemetry(telemetryInfo)
	query := req.URL.Query()
	query.Set("api-version", "2019-06-01")
	if options != nil && options.Maxpagesize != nil {
		query.Set("$maxpagesize", *options.Maxpagesize)
	}
	if options != nil && options.Filter != nil {
		query.Set("$filter", *options.Filter)
	}
	if options != nil && options.Expand != nil {
		query.Set("$expand", "deleted")
	}
	req.URL.RawQuery = query.Encode()
	req.Header.Set("Accept", "application/json")
	return req, nil
}

// listHandleResponse handles the List response.
func (client *FileSharesClient) listHandleResponse(resp *azcore.Response) (FileShareItemsResponse, error) {
	var val *FileShareItems
	if err := resp.UnmarshalAsJSON(&val); err != nil {
		return FileShareItemsResponse{}, err
	}
	return FileShareItemsResponse{RawResponse: resp.Response, FileShareItems: val}, nil
}

// listHandleError handles the List error response.
func (client *FileSharesClient) listHandleError(resp *azcore.Response) error {
	var err CloudError
	if err := resp.UnmarshalAsJSON(&err); err != nil {
		return err
	}
	return azcore.NewResponseError(&err, resp.Response)
}

// Restore - Restore a file share within a valid retention days if share soft delete is enabled
func (client *FileSharesClient) Restore(ctx context.Context, resourceGroupName string, accountName string, shareName string, deletedShare DeletedShare, options *FileSharesRestoreOptions) (*http.Response, error) {
	req, err := client.restoreCreateRequest(ctx, resourceGroupName, accountName, shareName, deletedShare, options)
	if err != nil {
		return nil, err
	}
	resp, err := client.con.Pipeline().Do(req)
	if err != nil {
		return nil, err
	}
	if !resp.HasStatusCode(http.StatusOK) {
		return nil, client.restoreHandleError(resp)
	}
	return resp.Response, nil
}

// restoreCreateRequest creates the Restore request.
func (client *FileSharesClient) restoreCreateRequest(ctx context.Context, resourceGroupName string, accountName string, shareName string, deletedShare DeletedShare, options *FileSharesRestoreOptions) (*azcore.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Storage/storageAccounts/{accountName}/fileServices/default/shares/{shareName}/restore"
	urlPath = strings.ReplaceAll(urlPath, "{resourceGroupName}", url.PathEscape(resourceGroupName))
	urlPath = strings.ReplaceAll(urlPath, "{accountName}", url.PathEscape(accountName))
	urlPath = strings.ReplaceAll(urlPath, "{shareName}", url.PathEscape(shareName))
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionId}", url.PathEscape(client.subscriptionID))
	req, err := azcore.NewRequest(ctx, http.MethodPost, azcore.JoinPaths(client.con.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	req.Telemetry(telemetryInfo)
	query := req.URL.Query()
	query.Set("api-version", "2019-06-01")
	req.URL.RawQuery = query.Encode()
	req.Header.Set("Accept", "application/json")
	return req, req.MarshalAsJSON(deletedShare)
}

// restoreHandleError handles the Restore error response.
func (client *FileSharesClient) restoreHandleError(resp *azcore.Response) error {
	var err CloudError
	if err := resp.UnmarshalAsJSON(&err); err != nil {
		return err
	}
	return azcore.NewResponseError(&err, resp.Response)
}

// Update - Updates share properties as specified in request body. Properties not mentioned in the request will not be changed. Update fails if the specified
// share does not already exist.
func (client *FileSharesClient) Update(ctx context.Context, resourceGroupName string, accountName string, shareName string, fileShare FileShare, options *FileSharesUpdateOptions) (FileShareResponse, error) {
	req, err := client.updateCreateRequest(ctx, resourceGroupName, accountName, shareName, fileShare, options)
	if err != nil {
		return FileShareResponse{}, err
	}
	resp, err := client.con.Pipeline().Do(req)
	if err != nil {
		return FileShareResponse{}, err
	}
	if !resp.HasStatusCode(http.StatusOK) {
		return FileShareResponse{}, client.updateHandleError(resp)
	}
	return client.updateHandleResponse(resp)
}

// updateCreateRequest creates the Update request.
func (client *FileSharesClient) updateCreateRequest(ctx context.Context, resourceGroupName string, accountName string, shareName string, fileShare FileShare, options *FileSharesUpdateOptions) (*azcore.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Storage/storageAccounts/{accountName}/fileServices/default/shares/{shareName}"
	urlPath = strings.ReplaceAll(urlPath, "{resourceGroupName}", url.PathEscape(resourceGroupName))
	urlPath = strings.ReplaceAll(urlPath, "{accountName}", url.PathEscape(accountName))
	urlPath = strings.ReplaceAll(urlPath, "{shareName}", url.PathEscape(shareName))
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionId}", url.PathEscape(client.subscriptionID))
	req, err := azcore.NewRequest(ctx, http.MethodPatch, azcore.JoinPaths(client.con.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	req.Telemetry(telemetryInfo)
	query := req.URL.Query()
	query.Set("api-version", "2019-06-01")
	req.URL.RawQuery = query.Encode()
	req.Header.Set("Accept", "application/json")
	return req, req.MarshalAsJSON(fileShare)
}

// updateHandleResponse handles the Update response.
func (client *FileSharesClient) updateHandleResponse(resp *azcore.Response) (FileShareResponse, error) {
	var val *FileShare
	if err := resp.UnmarshalAsJSON(&val); err != nil {
		return FileShareResponse{}, err
	}
	return FileShareResponse{RawResponse: resp.Response, FileShare: val}, nil
}

// updateHandleError handles the Update error response.
func (client *FileSharesClient) updateHandleError(resp *azcore.Response) error {
	var err CloudError
	if err := resp.UnmarshalAsJSON(&err); err != nil {
		return err
	}
	return azcore.NewResponseError(&err, resp.Response)
}
