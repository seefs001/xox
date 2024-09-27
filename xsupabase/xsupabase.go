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
	"github.com/seefs001/xox/xerror"
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
	httpClient := x.Must1(xhttpc.NewClient(append(options,
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
	)...))

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
	Message string      `json:"message"`
	Code    interface{} `json:"code"`
	Details string      `json:"details,omitempty"`
}

// User represents a user in the Supabase auth system
type User struct {
	ID               string    `json:"id"`
	Aud              string    `json:"aud"`
	Role             string    `json:"role"`
	Email            string    `json:"email"`
	EmailConfirmedAt time.Time `json:"email_confirmed_at"`
	Phone            string    `json:"phone"`
	ConfirmedAt      time.Time `json:"confirmed_at"`
	LastSignInAt     time.Time `json:"last_sign_in_at"`
	AppMetadata      Record    `json:"app_metadata"`
	UserMetadata     Record    `json:"user_metadata"`
	IdentityData     Record    `json:"identity_data"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// SetDebug enables or disables debug mode
func (c *Client) SetDebug(debug bool) {
	c.debug = debug
	c.httpClient.SetDebug(debug)
}

// SupabaseResponse represents a response from the Supabase API
type SupabaseResponse struct {
	Data   json.RawMessage `json:"data"`
	Error  *ErrorResponse  `json:"error,omitempty"`
	Status int             `json:"status"`
}

// execute sends an HTTP request to the Supabase API
func (c *Client) execute(ctx context.Context, method, path string, body interface{}) (*SupabaseResponse, error) {
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
			return xerror.NewErrorf("unsupported HTTP method: %s", method)
		}
		return err
	}

	err = operation()
	if err != nil {
		return nil, xerror.Wrap(err, "error executing request")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, xerror.Wrap(err, "error reading response body")
	}

	supaResp := &SupabaseResponse{
		Status: resp.StatusCode,
		Data:   respBody,
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return supaResp, xerror.Wrapf(err, "error unmarshaling error response (body: %s)", string(respBody))
		}
		supaResp.Error = &errResp

		codeStr := fmt.Sprintf("%v", errResp.Code)

		return supaResp, xerror.NewWithCode(fmt.Sprintf("API error: %s (code: %s, details: %s)", errResp.Message, codeStr, errResp.Details), resp.StatusCode)
	}

	return supaResp, nil
}

// BuildQueryString constructs a query string from QueryParams
func BuildQueryString(params QueryParams) string {
	var queryParts []string

	if params.Select != "" {
		queryParts = append(queryParts, "select="+params.Select)
	}
	if params.Order != "" {
		queryParts = append(queryParts, "order="+params.Order)
	}
	if params.Limit > 0 {
		queryParts = append(queryParts, "limit="+strconv.Itoa(params.Limit))
	}
	if params.Offset > 0 {
		queryParts = append(queryParts, "offset="+strconv.Itoa(params.Offset))
	}
	if params.Filter != "" {
		queryParts = append(queryParts, params.Filter)
	}

	return strings.Join(queryParts, "&")
}

// Select retrieves records from the specified table
func (c *Client) Select(ctx context.Context, table string, params QueryParams) ([]Record, error) {
	path := fmt.Sprintf("/rest/v1/%s", table)
	query := BuildQueryString(params)

	if len(query) > 0 {
		path += "?" + query
	}

	supaResp, err := c.execute(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, xerror.Wrap(err, "Select operation failed")
	}

	var records []Record
	if err := json.Unmarshal(supaResp.Data, &records); err != nil {
		return nil, xerror.Wrapf(err, "error unmarshaling response (body: %s)", string(supaResp.Data))
	}

	if c.debug {
		xlog.Info("Select operation", "table", table, "params", params, "records_count", len(records), "status", supaResp.Status)
	}

	return records, nil
}

// Insert inserts a new record into the specified table
func (c *Client) Insert(ctx context.Context, table string, record Record) (Record, error) {
	path := fmt.Sprintf("/rest/v1/%s", table)

	supaResp, err := c.execute(ctx, http.MethodPost, path, record)
	if err != nil {
		return nil, xerror.Wrap(err, "Insert operation failed")
	}

	if len(supaResp.Data) == 0 {
		if c.debug {
			xlog.Info("Insert operation - empty response", "table", table, "record", record, "status", supaResp.Status)
		}
		return record, nil
	}

	var insertedRecord Record
	if err := json.Unmarshal(supaResp.Data, &insertedRecord); err != nil {
		return nil, xerror.Wrapf(err, "error unmarshaling response (body: %s)", string(supaResp.Data))
	}

	if c.debug {
		xlog.Info("Insert operation", "table", table, "record", record, "inserted_record", insertedRecord, "status", supaResp.Status)
	}

	return insertedRecord, nil
}

// Update updates a record in the specified table
func (c *Client) Update(ctx context.Context, table string, id interface{}, record Record) (Record, error) {
	path := fmt.Sprintf("/rest/v1/%s?id=eq.%v", table, id)

	supaResp, err := c.execute(ctx, http.MethodPatch, path, record)
	if err != nil {
		return nil, xerror.Wrap(err, "Update operation failed")
	}

	if len(supaResp.Data) == 0 {
		if c.debug {
			xlog.Info("Update operation - empty response", "table", table, "id", id, "record", record, "status", supaResp.Status)
		}
		return record, nil
	}

	var updatedRecord Record
	if err := json.Unmarshal(supaResp.Data, &updatedRecord); err != nil {
		return nil, xerror.Wrapf(err, "error unmarshaling response (body: %s)", string(supaResp.Data))
	}

	if c.debug {
		xlog.Info("Update operation", "table", table, "id", id, "record", record, "updated_record", updatedRecord, "status", supaResp.Status)
	}

	return updatedRecord, nil
}

// Delete deletes a record from the specified table
func (c *Client) Delete(ctx context.Context, table string, id interface{}) error {
	path := fmt.Sprintf("/rest/v1/%s?id=eq.%v", table, id)

	supaResp, err := c.execute(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return xerror.Wrap(err, "Delete operation failed")
	}

	if len(supaResp.Data) == 0 {
		if c.debug {
			xlog.Info("Delete operation - empty response", "table", table, "id", id, "status", supaResp.Status)
		}
		return nil
	}

	if c.debug {
		xlog.Info("Delete operation", "table", table, "id", id, "response", string(supaResp.Data), "status", supaResp.Status)
	}

	return nil
}

// ExecuteRPC executes a stored procedure or function
func (c *Client) ExecuteRPC(ctx context.Context, functionName string, params map[string]interface{}) (json.RawMessage, error) {
	path := fmt.Sprintf("/rest/v1/rpc/%s", functionName)

	supaResp, err := c.execute(ctx, http.MethodPost, path, params)
	if err != nil {
		return nil, xerror.Wrap(err, "ExecuteRPC operation failed")
	}

	if c.debug {
		xlog.Info("ExecuteRPC operation", "function", functionName, "params", params, "response", string(supaResp.Data), "status", supaResp.Status)
	}

	return supaResp.Data, nil
}

// Count returns the number of records in the specified table
func (c *Client) Count(ctx context.Context, table string, filter string) (int, error) {
	path := fmt.Sprintf("/rest/v1/%s?select=count", table)
	if filter != "" {
		path += "&" + filter
	}

	supaResp, err := c.execute(ctx, http.MethodGet, path, nil)
	if err != nil {
		return 0, xerror.Wrap(err, "Count operation failed")
	}

	var result []struct {
		Count int `json:"count"`
	}
	if err := json.Unmarshal(supaResp.Data, &result); err != nil {
		return 0, xerror.Wrapf(err, "error unmarshaling response (body: %s)", string(supaResp.Data))
	}

	if len(result) == 0 {
		return 0, xerror.New("unexpected empty response")
	}

	count := result[0].Count

	if c.debug {
		xlog.Info("Count operation", "table", table, "filter", filter, "count", count, "status", supaResp.Status)
	}

	return count, nil
}

// Upsert inserts or updates records in the specified table
func (c *Client) Upsert(ctx context.Context, table string, records []Record, onConflict string) ([]Record, error) {
	path := fmt.Sprintf("/rest/v1/%s", table)
	if onConflict != "" {
		path += "?on_conflict=" + onConflict
	}

	supaResp, err := c.execute(ctx, http.MethodPost, path, records)
	if err != nil {
		return nil, xerror.Wrap(err, "Upsert operation failed")
	}

	var upsertedRecords []Record
	if err := json.Unmarshal(supaResp.Data, &upsertedRecords); err != nil {
		return nil, xerror.Wrapf(err, "error unmarshaling response (body: %s)", string(supaResp.Data))
	}

	if c.debug {
		xlog.Info("Upsert operation", "table", table, "records_count", len(records), "on_conflict", onConflict, "upserted_records_count", len(upsertedRecords), "status", supaResp.Status)
	}

	return upsertedRecords, nil
}

// BatchOperation performs a batch operation on the specified table
func (c *Client) BatchOperation(ctx context.Context, table string, operations []map[string]interface{}) ([]Record, error) {
	path := fmt.Sprintf("/rest/v1/%s", table)

	supaResp, err := c.execute(ctx, http.MethodPost, path, operations)
	if err != nil {
		return nil, xerror.Wrap(err, "BatchOperation failed")
	}

	var results []Record
	if err := json.Unmarshal(supaResp.Data, &results); err != nil {
		return nil, xerror.Wrapf(err, "error unmarshaling response (body: %s)", string(supaResp.Data))
	}

	if c.debug {
		xlog.Info("BatchOperation", "table", table, "operations_count", len(operations), "results_count", len(results), "status", supaResp.Status)
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

// CreateUser creates a new user
func (c *Client) CreateUser(ctx context.Context, email, password string, userData Record) (*User, error) {
	path := "/auth/v1/admin/users"
	body := map[string]interface{}{
		"email":    email,
		"password": password,
		"data":     userData,
	}

	supaResp, err := c.execute(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, xerror.Wrap(err, "CreateUser operation failed")
	}

	var user User
	if err := json.Unmarshal(supaResp.Data, &user); err != nil {
		return nil, xerror.Wrapf(err, "error unmarshaling response (body: %s)", string(supaResp.Data))
	}

	if c.debug {
		xlog.Info("CreateUser operation", "email", email, "user_id", user.ID, "status", supaResp.Status)
	}

	return &user, nil
}

// GetUser retrieves a user by ID
func (c *Client) GetUser(ctx context.Context, userID string) (*User, error) {
	path := fmt.Sprintf("/auth/v1/admin/users/%s", userID)

	supaResp, err := c.execute(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, xerror.Wrap(err, "GetUser operation failed")
	}

	var user User
	if err := json.Unmarshal(supaResp.Data, &user); err != nil {
		return nil, xerror.Wrapf(err, "error unmarshaling response (body: %s)", string(supaResp.Data))
	}

	if c.debug {
		xlog.Info("GetUser operation", "user_id", userID, "status", supaResp.Status)
	}

	return &user, nil
}

// UpdateUser updates a user's information
func (c *Client) UpdateUser(ctx context.Context, userID string, updates Record) (*User, error) {
	path := fmt.Sprintf("/auth/v1/admin/users/%s", userID)

	supaResp, err := c.execute(ctx, http.MethodPut, path, updates)
	if err != nil {
		return nil, xerror.Wrap(err, "UpdateUser operation failed")
	}

	var user User
	if err := json.Unmarshal(supaResp.Data, &user); err != nil {
		return nil, xerror.Wrapf(err, "error unmarshaling response (body: %s)", string(supaResp.Data))
	}

	if c.debug {
		xlog.Info("UpdateUser operation", "user_id", userID, "updates", updates, "status", supaResp.Status)
	}

	return &user, nil
}

// DeleteUser deletes a user
func (c *Client) DeleteUser(ctx context.Context, userID string) error {
	path := fmt.Sprintf("/auth/v1/admin/users/%s", userID)

	supaResp, err := c.execute(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return xerror.Wrap(err, "DeleteUser operation failed")
	}

	if c.debug {
		xlog.Info("DeleteUser operation", "user_id", userID, "status", supaResp.Status)
	}

	return nil
}

// ListUsers retrieves a list of users
func (c *Client) ListUsers(ctx context.Context, page, perPage int) ([]User, error) {
	path := fmt.Sprintf("/auth/v1/admin/users?page=%d&per_page=%d", page, perPage)

	supaResp, err := c.execute(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, xerror.Wrap(err, "ListUsers operation failed")
	}

	var users []User
	if err := json.Unmarshal(supaResp.Data, &users); err != nil {
		return nil, xerror.Wrapf(err, "error unmarshaling response (body: %s)", string(supaResp.Data))
	}

	if c.debug {
		xlog.Info("ListUsers operation", "page", page, "per_page", perPage, "users_count", len(users), "status", supaResp.Status)
	}

	return users, nil
}
