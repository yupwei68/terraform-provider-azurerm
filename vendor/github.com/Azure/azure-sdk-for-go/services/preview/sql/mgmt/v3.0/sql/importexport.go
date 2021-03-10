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
	"github.com/Azure/go-autorest/autorest/validation"
	"github.com/Azure/go-autorest/tracing"
	"net/http"
)

// ImportExportClient is the the Azure SQL Database management API provides a RESTful set of web services that interact
// with Azure SQL Database services to manage your databases. The API enables you to create, retrieve, update, and
// delete databases.
type ImportExportClient struct {
	BaseClient
}

// NewImportExportClient creates an instance of the ImportExportClient client.
func NewImportExportClient(subscriptionID string) ImportExportClient {
	return NewImportExportClientWithBaseURI(DefaultBaseURI, subscriptionID)
}

// NewImportExportClientWithBaseURI creates an instance of the ImportExportClient client using a custom endpoint.  Use
// this when interacting with an Azure cloud that uses a non-standard base URI (sovereign clouds, Azure stack).
func NewImportExportClientWithBaseURI(baseURI string, subscriptionID string) ImportExportClient {
	return ImportExportClient{NewWithBaseURI(baseURI, subscriptionID)}
}

// Import imports a bacpac into a new database.
// Parameters:
// resourceGroupName - the name of the resource group that contains the resource. You can obtain this value
// from the Azure Resource Manager API or the portal.
// serverName - the name of the server.
// databaseName - the name of the database.
// parameters - the database import request parameters.
func (client ImportExportClient) Import(ctx context.Context, resourceGroupName string, serverName string, databaseName string, parameters ImportExistingDatabaseDefinition) (result ImportExportImportFuture, err error) {
	if tracing.IsEnabled() {
		ctx = tracing.StartSpan(ctx, fqdn+"/ImportExportClient.Import")
		defer func() {
			sc := -1
			if result.FutureAPI != nil && result.FutureAPI.Response() != nil {
				sc = result.FutureAPI.Response().StatusCode
			}
			tracing.EndSpan(ctx, sc, err)
		}()
	}
	if err := validation.Validate([]validation.Validation{
		{TargetValue: parameters,
			Constraints: []validation.Constraint{{Target: "parameters.StorageKey", Name: validation.Null, Rule: true, Chain: nil},
				{Target: "parameters.StorageURI", Name: validation.Null, Rule: true, Chain: nil},
				{Target: "parameters.AdministratorLogin", Name: validation.Null, Rule: true, Chain: nil},
				{Target: "parameters.AdministratorLoginPassword", Name: validation.Null, Rule: true, Chain: nil}}}}); err != nil {
		return result, validation.NewError("sql.ImportExportClient", "Import", err.Error())
	}

	req, err := client.ImportPreparer(ctx, resourceGroupName, serverName, databaseName, parameters)
	if err != nil {
		err = autorest.NewErrorWithError(err, "sql.ImportExportClient", "Import", nil, "Failure preparing request")
		return
	}

	result, err = client.ImportSender(req)
	if err != nil {
		err = autorest.NewErrorWithError(err, "sql.ImportExportClient", "Import", nil, "Failure sending request")
		return
	}

	return
}

// ImportPreparer prepares the Import request.
func (client ImportExportClient) ImportPreparer(ctx context.Context, resourceGroupName string, serverName string, databaseName string, parameters ImportExistingDatabaseDefinition) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"databaseName":      autorest.Encode("path", databaseName),
		"resourceGroupName": autorest.Encode("path", resourceGroupName),
		"serverName":        autorest.Encode("path", serverName),
		"subscriptionId":    autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2020-02-02-preview"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsPost(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Sql/servers/{serverName}/databases/{databaseName}/import", pathParameters),
		autorest.WithJSON(parameters),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// ImportSender sends the Import request. The method will close the
// http.Response Body if it receives an error.
func (client ImportExportClient) ImportSender(req *http.Request) (future ImportExportImportFuture, err error) {
	var resp *http.Response
	resp, err = client.Send(req, azure.DoRetryWithRegistration(client.Client))
	if err != nil {
		return
	}
	var azf azure.Future
	azf, err = azure.NewFutureFromResponse(resp)
	future.FutureAPI = &azf
	future.Result = func(client ImportExportClient) (ieor ImportExportOperationResult, err error) {
		var done bool
		done, err = future.DoneWithContext(context.Background(), client)
		if err != nil {
			err = autorest.NewErrorWithError(err, "sql.ImportExportImportFuture", "Result", future.Response(), "Polling failure")
			return
		}
		if !done {
			err = azure.NewAsyncOpIncompleteError("sql.ImportExportImportFuture")
			return
		}
		sender := autorest.DecorateSender(client, autorest.DoRetryForStatusCodes(client.RetryAttempts, client.RetryDuration, autorest.StatusCodesForRetry...))
		ieor.Response.Response, err = future.GetResult(sender)
		if ieor.Response.Response == nil && err == nil {
			err = autorest.NewErrorWithError(err, "sql.ImportExportImportFuture", "Result", nil, "received nil response and error")
		}
		if err == nil && ieor.Response.Response.StatusCode != http.StatusNoContent {
			ieor, err = client.ImportResponder(ieor.Response.Response)
			if err != nil {
				err = autorest.NewErrorWithError(err, "sql.ImportExportImportFuture", "Result", ieor.Response.Response, "Failure responding to request")
			}
		}
		return
	}
	return
}

// ImportResponder handles the response to the Import request. The method always
// closes the http.Response Body.
func (client ImportExportClient) ImportResponder(resp *http.Response) (result ImportExportOperationResult, err error) {
	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusAccepted),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}
