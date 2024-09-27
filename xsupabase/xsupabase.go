package xsupabase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xhttpc"
	"github.com/seefs001/xox/xlog"
)

// Client represents a Supabase client
type Client struct {
	projectURL string
	apiKey     string
	httpClient *xhttpc.Client
	debug      bool
}

// NewClient creates a new Supabase client
func NewClient(projectURL, apiKey string, options ...xhttpc.ClientOption) *Client {
	httpClient := xhttpc.NewClient(append(options,
		xhttpc.WithTimeout(30*time.Second),
		xhttpc.WithRetryConfig(xhttpc.RetryConfig{
			Enabled:    true,
			Count:      3,
			MaxBackoff: 5 * time.Second,
		}),
		xhttpc.WithLogOptions(xhttpc.LogOptions{
			LogHeaders:     true,
			LogBody:        true,
			LogResponse:    true,
			MaxBodyLogSize: 1024,
		}),
	)...)

	return &Client{
		projectURL: strings.TrimRight(projectURL, "/"),
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

// Record represents a generic database record
type Record map[string]interface{}

// QueryParams represents query parameters for database operations
type QueryParams struct {
	Select string
	Order  string
	Limit  int
	Offset int
	Filter string
}

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

// SetDebug enables or disables debug mode
func (c *Client) SetDebug(debug bool) {
	c.debug = debug
	c.httpClient.SetDebug(debug)
}

// execute sends an HTTP request to the Supabase API
func (c *Client) execute(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.projectURL, path)

	c.httpClient.SetHeader("apikey", c.apiKey)
	c.httpClient.SetHeader("Authorization", "Bearer "+c.apiKey)
	c.httpClient.SetHeader("Content-Type", "application/json")

	var resp *http.Response
	var err error

	operation := func() error {
		switch method {
		case http.MethodGet:
			resp, err = c.httpClient.Get(ctx, url)
		case http.MethodPost:
			resp, err = c.httpClient.PostJSON(ctx, url, body)
		case http.MethodPatch:
			resp, err = c.httpClient.PatchJSON(ctx, url, body)
		case http.MethodDelete:
			resp, err = c.httpClient.Delete(ctx, url)
		default:
			return fmt.Errorf("unsupported HTTP method: %s", method)
		}
		return err
	}

	err = operation()
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return nil, fmt.Errorf("error unmarshaling error response: %w", err)
		}
		return nil, fmt.Errorf("API error: %s (code: %s)", errResp.Message, errResp.Code)
	}

	return respBody, nil
}

// Select retrieves records from the specified table
func (c *Client) Select(ctx context.Context, table string, params QueryParams) ([]Record, error) {
	path := fmt.Sprintf("/rest/v1/%s", table)
	query := make([]string, 0)

	if params.Select != "" {
		query = append(query, "select="+params.Select)
	}
	if params.Order != "" {
		query = append(query, "order="+params.Order)
	}
	if params.Limit > 0 {
		query = append(query, "limit="+strconv.Itoa(params.Limit))
	}
	if params.Offset > 0 {
		query = append(query, "offset="+strconv.Itoa(params.Offset))
	}
	if params.Filter != "" {
		query = append(query, params.Filter)
	}

	if len(query) > 0 {
		path += "?" + strings.Join(query, "&")
	}

	respBody, err := c.execute(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var records []Record
	if err := json.Unmarshal(respBody, &records); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	if c.debug {
		xlog.Info("Select operation", "table", table, "params", params, "records_count", len(records))
	}

	return records, nil
}

// Insert inserts a new record into the specified table
func (c *Client) Insert(ctx context.Context, table string, record Record) (Record, error) {
	path := fmt.Sprintf("/rest/v1/%s", table)

	respBody, err := c.execute(ctx, http.MethodPost, path, record)
	if err != nil {
		return nil, err
	}

	// Handle empty response
	if len(respBody) == 0 {
		if c.debug {
			xlog.Info("Insert operation - empty response", "table", table, "record", record)
		}
		return record, nil
	}

	var insertedRecord Record
	if err := json.Unmarshal(respBody, &insertedRecord); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	if c.debug {
		xlog.Info("Insert operation", "table", table, "record", record, "inserted_record", insertedRecord)
	}

	return insertedRecord, nil
}

// Update updates a record in the specified table
func (c *Client) Update(ctx context.Context, table string, id interface{}, record Record) (Record, error) {
	path := fmt.Sprintf("/rest/v1/%s?id=eq.%v", table, id)

	respBody, err := c.execute(ctx, http.MethodPatch, path, record)
	if err != nil {
		return nil, err
	}

	// Handle empty response for successful update
	if len(respBody) == 0 {
		if c.debug {
			xlog.Info("Update operation - empty response", "table", table, "id", id, "record", record)
		}
		// Return the original record as Supabase doesn't return the updated record
		return record, nil
	}

	var updatedRecord Record
	if err := json.Unmarshal(respBody, &updatedRecord); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	if c.debug {
		xlog.Info("Update operation", "table", table, "id", id, "record", record, "updated_record", updatedRecord)
	}

	return updatedRecord, nil
}

// Delete deletes a record from the specified table
func (c *Client) Delete(ctx context.Context, table string, id interface{}) error {
	path := fmt.Sprintf("/rest/v1/%s?id=eq.%v", table, id)

	respBody, err := c.execute(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	// Handle empty response for successful delete
	if len(respBody) == 0 {
		if c.debug {
			xlog.Info("Delete operation - empty response", "table", table, "id", id)
		}
		return nil
	}

	if c.debug {
		xlog.Info("Delete operation", "table", table, "id", id, "response", string(respBody))
	}

	return nil
}

// ExecuteRPC executes a stored procedure or function
func (c *Client) ExecuteRPC(ctx context.Context, functionName string, params map[string]interface{}) (json.RawMessage, error) {
	path := fmt.Sprintf("/rest/v1/rpc/%s", functionName)

	respBody, err := c.execute(ctx, http.MethodPost, path, params)
	if err != nil {
		return nil, err
	}

	if c.debug {
		xlog.Info("ExecuteRPC operation", "function", functionName, "params", params, "response", string(respBody))
	}

	return json.RawMessage(respBody), nil
}

// Count returns the number of records in the specified table
func (c *Client) Count(ctx context.Context, table string, filter string) (int, error) {
	path := fmt.Sprintf("/rest/v1/%s?select=count", table)
	if filter != "" {
		path += "&" + filter
	}

	respBody, err := c.execute(ctx, http.MethodGet, path, nil)
	if err != nil {
		return 0, err
	}

	var result []struct {
		Count int `json:"count"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return 0, fmt.Errorf("error unmarshaling response: %w", err)
	}

	if len(result) == 0 {
		return 0, fmt.Errorf("unexpected empty response")
	}

	count := result[0].Count

	if c.debug {
		xlog.Info("Count operation", "table", table, "filter", filter, "count", count)
	}

	return count, nil
}

// Upsert inserts or updates records in the specified table
func (c *Client) Upsert(ctx context.Context, table string, records []Record, onConflict string) ([]Record, error) {
	path := fmt.Sprintf("/rest/v1/%s", table)
	if onConflict != "" {
		path += "?on_conflict=" + onConflict
	}

	respBody, err := c.execute(ctx, http.MethodPost, path, records)
	if err != nil {
		return nil, err
	}

	var upsertedRecords []Record
	if err := json.Unmarshal(respBody, &upsertedRecords); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	if c.debug {
		xlog.Info("Upsert operation", "table", table, "records_count", len(records), "on_conflict", onConflict, "upserted_records_count", len(upsertedRecords))
	}

	return upsertedRecords, nil
}

// BatchOperation performs a batch operation on the specified table
func (c *Client) BatchOperation(ctx context.Context, table string, operations []map[string]interface{}) ([]Record, error) {
	path := fmt.Sprintf("/rest/v1/%s", table)

	respBody, err := c.execute(ctx, http.MethodPost, path, operations)
	if err != nil {
		return nil, err
	}

	var results []Record
	if err := json.Unmarshal(respBody, &results); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	if c.debug {
		xlog.Info("BatchOperation", "table", table, "operations_count", len(operations), "results_count", len(results))
	}

	return results, nil
}

// AsyncSelect performs an asynchronous select operation
func (c *Client) AsyncSelect(ctx context.Context, table string, params QueryParams) *x.AsyncTask[[]Record] {
	return x.NewAsyncTask(func() ([]Record, error) {
		return c.Select(ctx, table, params)
	})
}

// AsyncInsert performs an asynchronous insert operation
func (c *Client) AsyncInsert(ctx context.Context, table string, record Record) *x.AsyncTask[Record] {
	return x.NewAsyncTask(func() (Record, error) {
		return c.Insert(ctx, table, record)
	})
}

// AsyncUpdate performs an asynchronous update operation
func (c *Client) AsyncUpdate(ctx context.Context, table string, id interface{}, record Record) *x.AsyncTask[Record] {
	return x.NewAsyncTask(func() (Record, error) {
		return c.Update(ctx, table, id, record)
	})
}

// AsyncDelete performs an asynchronous delete operation
func (c *Client) AsyncDelete(ctx context.Context, table string, id interface{}) *x.AsyncTask[struct{}] {
	return x.NewAsyncTask(func() (struct{}, error) {
		err := c.Delete(ctx, table, id)
		return struct{}{}, err
	})
}
