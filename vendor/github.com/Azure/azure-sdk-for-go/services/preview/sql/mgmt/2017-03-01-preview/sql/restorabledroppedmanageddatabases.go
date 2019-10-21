package sql

// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"context"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/tracing"
	"net/http"
)

// RestorableDroppedManagedDatabasesClient is the the Azure SQL Database management API provides a RESTful set of web
// services that interact with Azure SQL Database services to manage your databases. The API enables you to create,
// retrieve, update, and delete databases.
type RestorableDroppedManagedDatabasesClient struct {
	BaseClient
}

// NewRestorableDroppedManagedDatabasesClient creates an instance of the RestorableDroppedManagedDatabasesClient
// client.
func NewRestorableDroppedManagedDatabasesClient(subscriptionID string) RestorableDroppedManagedDatabasesClient {
	return NewRestorableDroppedManagedDatabasesClientWithBaseURI(DefaultBaseURI, subscriptionID)
}

// NewRestorableDroppedManagedDatabasesClientWithBaseURI creates an instance of the
// RestorableDroppedManagedDatabasesClient client.
func NewRestorableDroppedManagedDatabasesClientWithBaseURI(baseURI string, subscriptionID string) RestorableDroppedManagedDatabasesClient {
	return RestorableDroppedManagedDatabasesClient{NewWithBaseURI(baseURI, subscriptionID)}
}

// Get gets a restorable dropped managed database.
// Parameters:
// resourceGroupName - the name of the resource group that contains the resource. You can obtain this value
// from the Azure Resource Manager API or the portal.
// managedInstanceName - the name of the managed instance.
func (client RestorableDroppedManagedDatabasesClient) Get(ctx context.Context, resourceGroupName string, managedInstanceName string, restorableDroppedDatabaseID string) (result RestorableDroppedManagedDatabase, err error) {
	if tracing.IsEnabled() {
		ctx = tracing.StartSpan(ctx, fqdn+"/RestorableDroppedManagedDatabasesClient.Get")
		defer func() {
			sc := -1
			if result.Response.Response != nil {
				sc = result.Response.Response.StatusCode
			}
			tracing.EndSpan(ctx, sc, err)
		}()
	}
	req, err := client.GetPreparer(ctx, resourceGroupName, managedInstanceName, restorableDroppedDatabaseID)
	if err != nil {
		err = autorest.NewErrorWithError(err, "sql.RestorableDroppedManagedDatabasesClient", "Get", nil, "Failure preparing request")
		return
	}

	resp, err := client.GetSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "sql.RestorableDroppedManagedDatabasesClient", "Get", resp, "Failure sending request")
		return
	}

	result, err = client.GetResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "sql.RestorableDroppedManagedDatabasesClient", "Get", resp, "Failure responding to request")
	}

	return
}

// GetPreparer prepares the Get request.
func (client RestorableDroppedManagedDatabasesClient) GetPreparer(ctx context.Context, resourceGroupName string, managedInstanceName string, restorableDroppedDatabaseID string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"managedInstanceName":         autorest.Encode("path", managedInstanceName),
		"resourceGroupName":           autorest.Encode("path", resourceGroupName),
		"restorableDroppedDatabaseId": autorest.Encode("path", restorableDroppedDatabaseID),
		"subscriptionId":              autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2017-03-01-preview"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsGet(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Sql/managedInstances/{managedInstanceName}/restorableDroppedDatabases/{restorableDroppedDatabaseId}", pathParameters),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// GetSender sends the Get request. The method will close the
// http.Response Body if it receives an error.
func (client RestorableDroppedManagedDatabasesClient) GetSender(req *http.Request) (*http.Response, error) {
	sd := autorest.GetSendDecorators(req.Context(), azure.DoRetryWithRegistration(client.Client))
	return autorest.SendWithSender(client, req, sd...)
}

// GetResponder handles the response to the Get request. The method always
// closes the http.Response Body.
func (client RestorableDroppedManagedDatabasesClient) GetResponder(resp *http.Response) (result RestorableDroppedManagedDatabase, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// ListByInstance gets a list of restorable dropped managed databases.
// Parameters:
// resourceGroupName - the name of the resource group that contains the resource. You can obtain this value
// from the Azure Resource Manager API or the portal.
// managedInstanceName - the name of the managed instance.
func (client RestorableDroppedManagedDatabasesClient) ListByInstance(ctx context.Context, resourceGroupName string, managedInstanceName string) (result RestorableDroppedManagedDatabaseListResultPage, err error) {
	if tracing.IsEnabled() {
		ctx = tracing.StartSpan(ctx, fqdn+"/RestorableDroppedManagedDatabasesClient.ListByInstance")
		defer func() {
			sc := -1
			if result.rdmdlr.Response.Response != nil {
				sc = result.rdmdlr.Response.Response.StatusCode
			}
			tracing.EndSpan(ctx, sc, err)
		}()
	}
	result.fn = client.listByInstanceNextResults
	req, err := client.ListByInstancePreparer(ctx, resourceGroupName, managedInstanceName)
	if err != nil {
		err = autorest.NewErrorWithError(err, "sql.RestorableDroppedManagedDatabasesClient", "ListByInstance", nil, "Failure preparing request")
		return
	}

	resp, err := client.ListByInstanceSender(req)
	if err != nil {
		result.rdmdlr.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "sql.RestorableDroppedManagedDatabasesClient", "ListByInstance", resp, "Failure sending request")
		return
	}

	result.rdmdlr, err = client.ListByInstanceResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "sql.RestorableDroppedManagedDatabasesClient", "ListByInstance", resp, "Failure responding to request")
	}

	return
}

// ListByInstancePreparer prepares the ListByInstance request.
func (client RestorableDroppedManagedDatabasesClient) ListByInstancePreparer(ctx context.Context, resourceGroupName string, managedInstanceName string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"managedInstanceName": autorest.Encode("path", managedInstanceName),
		"resourceGroupName":   autorest.Encode("path", resourceGroupName),
		"subscriptionId":      autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2017-03-01-preview"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsGet(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Sql/managedInstances/{managedInstanceName}/restorableDroppedDatabases", pathParameters),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// ListByInstanceSender sends the ListByInstance request. The method will close the
// http.Response Body if it receives an error.
func (client RestorableDroppedManagedDatabasesClient) ListByInstanceSender(req *http.Request) (*http.Response, error) {
	sd := autorest.GetSendDecorators(req.Context(), azure.DoRetryWithRegistration(client.Client))
	return autorest.SendWithSender(client, req, sd...)
}

// ListByInstanceResponder handles the response to the ListByInstance request. The method always
// closes the http.Response Body.
func (client RestorableDroppedManagedDatabasesClient) ListByInstanceResponder(resp *http.Response) (result RestorableDroppedManagedDatabaseListResult, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// listByInstanceNextResults retrieves the next set of results, if any.
func (client RestorableDroppedManagedDatabasesClient) listByInstanceNextResults(ctx context.Context, lastResults RestorableDroppedManagedDatabaseListResult) (result RestorableDroppedManagedDatabaseListResult, err error) {
	req, err := lastResults.restorableDroppedManagedDatabaseListResultPreparer(ctx)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "sql.RestorableDroppedManagedDatabasesClient", "listByInstanceNextResults", nil, "Failure preparing next results request")
	}
	if req == nil {
		return
	}
	resp, err := client.ListByInstanceSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "sql.RestorableDroppedManagedDatabasesClient", "listByInstanceNextResults", resp, "Failure sending next results request")
	}
	result, err = client.ListByInstanceResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "sql.RestorableDroppedManagedDatabasesClient", "listByInstanceNextResults", resp, "Failure responding to next results request")
	}
	return
}

// ListByInstanceComplete enumerates all values, automatically crossing page boundaries as required.
func (client RestorableDroppedManagedDatabasesClient) ListByInstanceComplete(ctx context.Context, resourceGroupName string, managedInstanceName string) (result RestorableDroppedManagedDatabaseListResultIterator, err error) {
	if tracing.IsEnabled() {
		ctx = tracing.StartSpan(ctx, fqdn+"/RestorableDroppedManagedDatabasesClient.ListByInstance")
		defer func() {
			sc := -1
			if result.Response().Response.Response != nil {
				sc = result.page.Response().Response.Response.StatusCode
			}
			tracing.EndSpan(ctx, sc, err)
		}()
	}
	result.page, err = client.ListByInstance(ctx, resourceGroupName, managedInstanceName)
	return
}
