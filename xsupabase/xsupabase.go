package xsupabase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xerror"
	"github.com/seefs001/xox/xhttpc"
	"github.com/seefs001/xox/xlog"
)

const (
	defaultTimeout     = 30 * time.Second
	defaultRetryCount  = 3
	defaultMaxBackoff  = 5 * time.Second
	defaultMaxBodySize = 1024
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
	defaultOptions := []xhttpc.ClientOption{
		xhttpc.WithTimeout(defaultTimeout),
		xhttpc.WithRetryConfig(xhttpc.RetryConfig{
			Enabled:    true,
			Count:      defaultRetryCount,
			MaxBackoff: defaultMaxBackoff,
		}),
		xhttpc.WithLogOptions(xhttpc.LogOptions{
			LogBody:        true,
			LogResponse:    true,
			MaxBodyLogSize: defaultMaxBodySize,
		}),
	}

	// Append user-provided options after default options
	allOptions := append(defaultOptions, options...)

	httpClient := x.Must1(xhttpc.NewClient(allOptions...))

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
	Message   string      `json:"msg"`
	Code      interface{} `json:"code"`
	ErrorCode string      `json:"error_code"`
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

// SetDebug enables or disables debug mode for xsupabase
func (c *Client) SetDebug(debug bool) {
	c.debug = debug
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
			resp, err = c.httpClient.Post(ctx, url, body)
		case http.MethodPatch:
			resp, err = c.httpClient.Patch(ctx, url, body)
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

	var errResp ErrorResponse
	if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Code != nil {
		supaResp.Error = &errResp
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return supaResp, xerror.NewWithCode(fmt.Sprintf("API error: %s (code: %v, error_code: %s)", errResp.Message, errResp.Code, errResp.ErrorCode), resp.StatusCode)
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
		return nil, err
	}

	var records []Record
	if err := json.Unmarshal(supaResp.Data, &records); err != nil {
		return nil, xerror.Wrapf(err, "error unmarshaling response (body: %s)", string(supaResp.Data))
	}

	if c.debug {
		xlog.Debug("Select operation", "table", table, "params", params, "records_count", len(records), "status", supaResp.Status)
	}

	return records, nil
}

// Insert inserts a new record into the specified table
func (c *Client) Insert(ctx context.Context, table string, record Record) (Record, error) {
	path := fmt.Sprintf("/rest/v1/%s", table)

	supaResp, err := c.execute(ctx, http.MethodPost, path, record)
	if err != nil {
		return nil, err
	}

	if len(supaResp.Data) == 0 {
		if c.debug {
			xlog.Debug("Insert operation - empty response", "table", table, "record", record, "status", supaResp.Status)
		}
		return record, nil
	}

	var insertedRecord Record
	if err := json.Unmarshal(supaResp.Data, &insertedRecord); err != nil {
		return nil, xerror.Wrapf(err, "error unmarshaling response (body: %s)", string(supaResp.Data))
	}

	if c.debug {
		xlog.Debug("Insert operation", "table", table, "record", record, "inserted_record", insertedRecord, "status", supaResp.Status)
	}

	return insertedRecord, nil
}

// Update updates a record in the specified table
func (c *Client) Update(ctx context.Context, table string, id interface{}, record Record) (Record, error) {
	path := fmt.Sprintf("/rest/v1/%s?id=eq.%v", table, id)

	supaResp, err := c.execute(ctx, http.MethodPatch, path, record)
	if err != nil {
		return nil, err
	}

	if len(supaResp.Data) == 0 {
		if c.debug {
			xlog.Debug("Update operation - empty response", "table", table, "id", id, "record", record, "status", supaResp.Status)
		}
		return record, nil
	}

	var updatedRecord Record
	if err := json.Unmarshal(supaResp.Data, &updatedRecord); err != nil {
		return nil, xerror.Wrapf(err, "error unmarshaling response (body: %s)", string(supaResp.Data))
	}

	if c.debug {
		xlog.Debug("Update operation", "table", table, "id", id, "record", record, "updated_record", updatedRecord, "status", supaResp.Status)
	}

	return updatedRecord, nil
}

// Delete deletes a record from the specified table
func (c *Client) Delete(ctx context.Context, table string, id interface{}) error {
	path := fmt.Sprintf("/rest/v1/%s?id=eq.%v", table, id)

	supaResp, err := c.execute(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	if len(supaResp.Data) == 0 {
		if c.debug {
			xlog.Debug("Delete operation - empty response", "table", table, "id", id, "status", supaResp.Status)
		}
		return nil
	}

	if c.debug {
		xlog.Debug("Delete operation", "table", table, "id", id, "response", string(supaResp.Data), "status", supaResp.Status)
	}

	return nil
}

// ExecuteRPC executes a stored procedure or function
func (c *Client) ExecuteRPC(ctx context.Context, functionName string, params map[string]interface{}) (json.RawMessage, error) {
	path := fmt.Sprintf("/rest/v1/rpc/%s", functionName)

	supaResp, err := c.execute(ctx, http.MethodPost, path, params)
	if err != nil {
		return nil, err
	}

	if c.debug {
		xlog.Debug("ExecuteRPC operation", "function", functionName, "params", params, "response", string(supaResp.Data), "status", supaResp.Status)
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
		return 0, err
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
		xlog.Debug("Count operation", "table", table, "filter", filter, "count", count, "status", supaResp.Status)
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
		return nil, err
	}

	var upsertedRecords []Record
	if err := json.Unmarshal(supaResp.Data, &upsertedRecords); err != nil {
		return nil, xerror.Wrapf(err, "error unmarshaling response (body: %s)", string(supaResp.Data))
	}

	if c.debug {
		xlog.Debug("Upsert operation", "table", table, "records_count", len(records), "on_conflict", onConflict, "upserted_records_count", len(upsertedRecords), "status", supaResp.Status)
	}

	return upsertedRecords, nil
}

// BatchOperation performs a batch operation on the specified table
func (c *Client) BatchOperation(ctx context.Context, table string, operations []map[string]interface{}) ([]Record, error) {
	path := fmt.Sprintf("/rest/v1/%s", table)

	supaResp, err := c.execute(ctx, http.MethodPost, path, operations)
	if err != nil {
		return nil, err
	}

	var results []Record
	if err := json.Unmarshal(supaResp.Data, &results); err != nil {
		return nil, xerror.Wrapf(err, "error unmarshaling response (body: %s)", string(supaResp.Data))
	}

	if c.debug {
		xlog.Debug("BatchOperation", "table", table, "operations_count", len(operations), "results_count", len(results), "status", supaResp.Status)
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
func (c *Client) CreateUser(ctx context.Context, email, password string, options ...CreateUserOption) (*User, error) {
	path := "/auth/v1/admin/users"
	body := map[string]interface{}{
		"email":    email,
		"password": password,
	}

	for _, option := range options {
		option(body)
	}

	supaResp, err := c.execute(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal(supaResp.Data, &user); err != nil {
		return nil, xerror.Wrapf(err, "error unmarshaling response (body: %s)", string(supaResp.Data))
	}

	if c.debug {
		xlog.Debug("CreateUser operation", "email", email, "user_id", user.ID, "status", supaResp.Status)
	}

	return &user, nil
}

// CreateUserOption defines the options for creating a user
type CreateUserOption func(map[string]interface{})

// WithUserMetadata sets the user metadata
func WithUserMetadata(metadata map[string]interface{}) CreateUserOption {
	return func(body map[string]interface{}) {
		body["user_metadata"] = metadata
	}
}

// WithAppMetadata sets the app metadata
func WithAppMetadata(metadata map[string]interface{}) CreateUserOption {
	return func(body map[string]interface{}) {
		body["app_metadata"] = metadata
	}
}

// WithPhone sets the user's phone number
func WithPhone(phone string) CreateUserOption {
	return func(body map[string]interface{}) {
		body["phone"] = phone
	}
}

// WithEmailConfirmed sets whether the email is confirmed
func WithEmailConfirmed(confirmed bool) CreateUserOption {
	return func(body map[string]interface{}) {
		body["email_confirm"] = confirmed
	}
}

// WithPhoneConfirmed sets whether the phone is confirmed
func WithPhoneConfirmed(confirmed bool) CreateUserOption {
	return func(body map[string]interface{}) {
		body["phone_confirm"] = confirmed
	}
}

// GetUser retrieves a user by ID
func (c *Client) GetUser(ctx context.Context, userID string) (*User, error) {
	path := fmt.Sprintf("/auth/v1/admin/users/%s", userID)

	supaResp, err := c.execute(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal(supaResp.Data, &user); err != nil {
		return nil, xerror.Wrapf(err, "error unmarshaling response (body: %s)", string(supaResp.Data))
	}

	if c.debug {
		xlog.Debug("GetUser operation", "user_id", userID, "status", supaResp.Status)
	}

	return &user, nil
}

// UpdateUser updates a user's information
func (c *Client) UpdateUser(ctx context.Context, userID string, options ...UpdateUserOption) (*User, error) {
	path := fmt.Sprintf("/auth/v1/admin/users/%s", userID)
	updates := make(map[string]interface{})

	for _, option := range options {
		option(updates)
	}

	supaResp, err := c.execute(ctx, http.MethodPut, path, updates)
	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal(supaResp.Data, &user); err != nil {
		return nil, xerror.Wrapf(err, "error unmarshaling response (body: %s)", string(supaResp.Data))
	}

	if c.debug {
		xlog.Debug("UpdateUser operation", "user_id", userID, "updates", updates, "status", supaResp.Status)
	}

	return &user, nil
}

// UpdateUserOption defines the options for updating a user
type UpdateUserOption func(map[string]interface{})

// WithEmail updates the user's email
func WithEmail(email string) UpdateUserOption {
	return func(updates map[string]interface{}) {
		updates["email"] = email
	}
}

// WithPassword updates the user's password
func WithPassword(password string) UpdateUserOption {
	return func(updates map[string]interface{}) {
		updates["password"] = password
	}
}

// WithUserMetadataUpdate updates the user's metadata
func WithUserMetadataUpdate(metadata map[string]interface{}) UpdateUserOption {
	return func(updates map[string]interface{}) {
		updates["user_metadata"] = metadata
	}
}

// WithAppMetadataUpdate updates the app metadata
func WithAppMetadataUpdate(metadata map[string]interface{}) UpdateUserOption {
	return func(updates map[string]interface{}) {
		updates["app_metadata"] = metadata
	}
}

// DeleteUser deletes a user
func (c *Client) DeleteUser(ctx context.Context, userID string) error {
	path := fmt.Sprintf("/auth/v1/admin/users/%s", userID)

	supaResp, err := c.execute(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	if c.debug {
		xlog.Debug("DeleteUser operation", "user_id", userID, "status", supaResp.Status)
	}

	return nil
}

// ListUsers retrieves a list of users
func (c *Client) ListUsers(ctx context.Context, options ...ListUsersOption) ([]User, error) {
	path := "/auth/v1/admin/users"
	queryParams := make(map[string]string)

	for _, option := range options {
		option(queryParams)
	}

	if len(queryParams) > 0 {
		query := url.Values{}
		for key, value := range queryParams {
			query.Add(key, value)
		}
		path += "?" + query.Encode()
	}

	supaResp, err := c.execute(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var users []User
	if err := json.Unmarshal(supaResp.Data, &users); err != nil {
		return nil, xerror.Wrapf(err, "error unmarshaling response (body: %s)", string(supaResp.Data))
	}

	if c.debug {
		xlog.Debug("ListUsers operation", "users_count", len(users), "status", supaResp.Status)
	}

	return users, nil
}

// ListUsersOption defines the options for listing users
type ListUsersOption func(map[string]string)

// WithPage sets the page number for pagination
func WithPage(page int) ListUsersOption {
	return func(params map[string]string) {
		params["page"] = strconv.Itoa(page)
	}
}

// WithPerPage sets the number of users per page
func WithPerPage(perPage int) ListUsersOption {
	return func(params map[string]string) {
		params["per_page"] = strconv.Itoa(perPage)
	}
}

// WithUserMetadataFilter filters users based on metadata
func WithUserMetadataFilter(key, value string) ListUsersOption {
	return func(params map[string]string) {
		params["user_metadata["+key+"]"] = value
	}
}

// ExecuteGraphQL executes a GraphQL query or mutation
func (c *Client) ExecuteGraphQL(ctx context.Context, query string, variables map[string]interface{}) (json.RawMessage, error) {
	path := "/graphql/v1"
	body := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	supaResp, err := c.execute(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}

	if c.debug {
		xlog.Debug("ExecuteGraphQL operation", "query", query, "variables", variables, "status", supaResp.Status)
	}

	return supaResp.Data, nil
}

// SubscribeToChanges subscribes to real-time changes in the specified table
func (c *Client) SubscribeToChanges(ctx context.Context, table string, callback func(Record)) error {
	// Implementation of real-time subscriptions would require WebSocket support
	// This is a placeholder for future implementation
	return xerror.New("SubscribeToChanges is not implemented yet")
}

// GetStoragePublicURL generates a public URL for a file in storage
func (c *Client) GetStoragePublicURL(ctx context.Context, bucket, path string) (string, error) {
	apiPath := fmt.Sprintf("/storage/v1/object/public/%s/%s", bucket, path)
	return c.projectURL + apiPath, nil
}

// UploadFile uploads a file to storage
func (c *Client) UploadFile(ctx context.Context, bucket, path string, file io.Reader) error {
	apiPath := fmt.Sprintf("/storage/v1/object/%s/%s", bucket, path)

	supaResp, err := c.execute(ctx, http.MethodPost, apiPath, file)
	if err != nil {
		return err
	}

	if c.debug {
		xlog.Debug("UploadFile operation", "bucket", bucket, "path", path, "status", supaResp.Status)
	}

	return nil
}

// DeleteFile deletes a file from storage
func (c *Client) DeleteFile(ctx context.Context, bucket, path string) error {
	apiPath := fmt.Sprintf("/storage/v1/object/%s/%s", bucket, path)

	supaResp, err := c.execute(ctx, http.MethodDelete, apiPath, nil)
	if err != nil {
		return err
	}

	if c.debug {
		xlog.Debug("DeleteFile operation", "bucket", bucket, "path", path, "status", supaResp.Status)
	}

	return nil
}

// ListFiles lists files in a storage bucket
func (c *Client) ListFiles(ctx context.Context, bucket string, prefix string) ([]string, error) {
	apiPath := fmt.Sprintf("/storage/v1/object/list/%s", bucket)
	if prefix != "" {
		apiPath += "?prefix=" + url.QueryEscape(prefix)
	}

	supaResp, err := c.execute(ctx, http.MethodGet, apiPath, nil)
	if err != nil {
		return nil, err
	}

	var files []string
	if err := json.Unmarshal(supaResp.Data, &files); err != nil {
		return nil, xerror.Wrapf(err, "error unmarshaling response (body: %s)", string(supaResp.Data))
	}

	if c.debug {
		xlog.Debug("ListFiles operation", "bucket", bucket, "prefix", prefix, "files_count", len(files), "status", supaResp.Status)
	}

	return files, nil
}

// CreateBucket creates a new storage bucket
func (c *Client) CreateBucket(ctx context.Context, bucketName string, isPublic bool) error {
	apiPath := "/storage/v1/bucket"
	body := map[string]interface{}{
		"name":            bucketName,
		"public":          isPublic,
		"file_size_limit": 50 * 1024 * 1024, // 50MB default limit
	}

	supaResp, err := c.execute(ctx, http.MethodPost, apiPath, body)
	if err != nil {
		return err
	}

	if c.debug {
		xlog.Debug("CreateBucket operation", "bucket", bucketName, "public", isPublic, "status", supaResp.Status)
	}

	return nil
}

// DeleteBucket deletes a storage bucket
func (c *Client) DeleteBucket(ctx context.Context, bucketName string) error {
	apiPath := fmt.Sprintf("/storage/v1/bucket/%s", bucketName)

	supaResp, err := c.execute(ctx, http.MethodDelete, apiPath, nil)
	if err != nil {
		return err
	}

	if c.debug {
		xlog.Debug("DeleteBucket operation", "bucket", bucketName, "status", supaResp.Status)
	}

	return nil
}

// GetBucketDetails retrieves details of a storage bucket
func (c *Client) GetBucketDetails(ctx context.Context, bucketName string) (map[string]interface{}, error) {
	apiPath := fmt.Sprintf("/storage/v1/bucket/%s", bucketName)

	supaResp, err := c.execute(ctx, http.MethodGet, apiPath, nil)
	if err != nil {
		return nil, err
	}

	var details map[string]interface{}
	if err := json.Unmarshal(supaResp.Data, &details); err != nil {
		return nil, xerror.Wrapf(err, "error unmarshaling response (body: %s)", string(supaResp.Data))
	}

	if c.debug {
		xlog.Debug("GetBucketDetails operation", "bucket", bucketName, "status", supaResp.Status)
	}

	return details, nil
}

// UpdateBucketDetails updates details of a storage bucket
func (c *Client) UpdateBucketDetails(ctx context.Context, bucketName string, updates map[string]interface{}) error {
	apiPath := fmt.Sprintf("/storage/v1/bucket/%s", bucketName)

	supaResp, err := c.execute(ctx, http.MethodPut, apiPath, updates)
	if err != nil {
		return err
	}

	if c.debug {
		xlog.Debug("UpdateBucketDetails operation", "bucket", bucketName, "updates", updates, "status", supaResp.Status)
	}

	return nil
}

// InviteUserByEmail invites a user to the project by email
func (c *Client) InviteUserByEmail(ctx context.Context, email string, role string) error {
	apiPath := "/auth/v1/invite"
	body := map[string]interface{}{
		"email": email,
		"role":  role,
	}

	supaResp, err := c.execute(ctx, http.MethodPost, apiPath, body)
	if err != nil {
		return err
	}

	if c.debug {
		xlog.Debug("InviteUserByEmail operation", "email", email, "role", role, "status", supaResp.Status)
	}

	return nil
}

// GetProjectSettings retrieves the project settings
func (c *Client) GetProjectSettings(ctx context.Context) (map[string]interface{}, error) {
	apiPath := "/rest/v1/settings"

	supaResp, err := c.execute(ctx, http.MethodGet, apiPath, nil)
	if err != nil {
		return nil, err
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(supaResp.Data, &settings); err != nil {
		return nil, xerror.Wrapf(err, "error unmarshaling response (body: %s)", string(supaResp.Data))
	}

	if c.debug {
		xlog.Debug("GetProjectSettings operation", "status", supaResp.Status)
	}

	return settings, nil
}
