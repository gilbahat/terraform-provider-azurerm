package cosmosdb

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/go-azure-helpers/polling"
)

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See NOTICE.txt in the project root for license information.

type SqlResourcesMigrateSqlDatabaseToAutoscaleOperationResponse struct {
	Poller       polling.LongRunningPoller
	HttpResponse *http.Response
	Model        *ThroughputSettingsGetResults
}

// SqlResourcesMigrateSqlDatabaseToAutoscale ...
func (c CosmosDBClient) SqlResourcesMigrateSqlDatabaseToAutoscale(ctx context.Context, id SqlDatabaseId) (result SqlResourcesMigrateSqlDatabaseToAutoscaleOperationResponse, err error) {
	req, err := c.preparerForSqlResourcesMigrateSqlDatabaseToAutoscale(ctx, id)
	if err != nil {
		err = autorest.NewErrorWithError(err, "cosmosdb.CosmosDBClient", "SqlResourcesMigrateSqlDatabaseToAutoscale", nil, "Failure preparing request")
		return
	}

	result, err = c.senderForSqlResourcesMigrateSqlDatabaseToAutoscale(ctx, req)
	if err != nil {
		err = autorest.NewErrorWithError(err, "cosmosdb.CosmosDBClient", "SqlResourcesMigrateSqlDatabaseToAutoscale", result.HttpResponse, "Failure sending request")
		return
	}

	return
}

// SqlResourcesMigrateSqlDatabaseToAutoscaleThenPoll performs SqlResourcesMigrateSqlDatabaseToAutoscale then polls until it's completed
func (c CosmosDBClient) SqlResourcesMigrateSqlDatabaseToAutoscaleThenPoll(ctx context.Context, id SqlDatabaseId) error {
	result, err := c.SqlResourcesMigrateSqlDatabaseToAutoscale(ctx, id)
	if err != nil {
		return fmt.Errorf("performing SqlResourcesMigrateSqlDatabaseToAutoscale: %+v", err)
	}

	if err := result.Poller.PollUntilDone(); err != nil {
		return fmt.Errorf("polling after SqlResourcesMigrateSqlDatabaseToAutoscale: %+v", err)
	}

	return nil
}

// preparerForSqlResourcesMigrateSqlDatabaseToAutoscale prepares the SqlResourcesMigrateSqlDatabaseToAutoscale request.
func (c CosmosDBClient) preparerForSqlResourcesMigrateSqlDatabaseToAutoscale(ctx context.Context, id SqlDatabaseId) (*http.Request, error) {
	queryParameters := map[string]interface{}{
		"api-version": defaultApiVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsPost(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithPath(fmt.Sprintf("%s/throughputSettings/default/migrateToAutoscale", id.ID())),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// senderForSqlResourcesMigrateSqlDatabaseToAutoscale sends the SqlResourcesMigrateSqlDatabaseToAutoscale request. The method will close the
// http.Response Body if it receives an error.
func (c CosmosDBClient) senderForSqlResourcesMigrateSqlDatabaseToAutoscale(ctx context.Context, req *http.Request) (future SqlResourcesMigrateSqlDatabaseToAutoscaleOperationResponse, err error) {
	var resp *http.Response
	resp, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		return
	}

	future.Poller, err = polling.NewPollerFromResponse(ctx, resp, c.Client, req.Method)
	return
}
