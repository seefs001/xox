package xresend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/seefs001/xox/xerror"
	"github.com/seefs001/xox/xhttpc"
	"github.com/seefs001/xox/xlog"
)

const (
	baseURL = "https://api.resend.com"
)

// APIError represents an error returned by the Resend API
type APIError struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Name       string `json:"name"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("API error: %s (status code: %d, name: %s)", e.Message, e.StatusCode, e.Name)
}

// Client is the Resend API client
type Client struct {
	httpClient *xhttpc.Client
	debug      bool
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// WithDebug enables or disables debug mode for the Resend client
func WithDebug(debug bool) ClientOption {
	return func(c *Client) {
		c.debug = debug
	}
}

// NewClient creates a new Resend API client
func NewClient(apiKey string, options ...ClientOption) (*Client, error) {
	httpClient, err := xhttpc.NewClient(
		xhttpc.WithBaseURL(baseURL),
		xhttpc.WithBearerToken(apiKey),
	)
	if err != nil {
		return nil, err
	}

	client := &Client{
		httpClient: httpClient,
		debug:      false,
	}

	for _, option := range options {
		option(client)
	}

	if client.debug {
		httpClient.SetDebug(true)
		httpClient.SetLogOptions(xhttpc.LogOptions{
			LogHeaders:  true,
			LogBody:     true,
			LogResponse: true,
		})
	}

	return client, nil
}

// toJSONBody converts a struct to xhttpc.JSONBody
func toJSONBody(v interface{}) (xhttpc.JSONBody, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var body map[string]interface{}
	err = json.Unmarshal(data, &body)
	if err != nil {
		return nil, err
	}
	return xhttpc.JSONBody(body), nil
}

// handleAPIError checks if the error is an API error and returns it as an APIError
func handleAPIError(err error) error {
	var apiErr APIError
	if xerror.As(err, &apiErr) {
		return &apiErr
	}
	return err
}

// sendRequest is a helper function to send requests and handle responses
func (c *Client) sendRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var resp *http.Response
	var err error

	switch method {
	case http.MethodGet:
		resp, err = c.httpClient.Get(ctx, path)
	case http.MethodPost:
		resp, err = c.httpClient.Post(ctx, path, body)
	case http.MethodDelete:
		resp, err = c.httpClient.Delete(ctx, path)
	default:
		return nil, xerror.Newf("unsupported HTTP method: %s", method)
	}

	if err != nil {
		return nil, xerror.Wrap(err, "request failed")
	}
	if resp == nil {
		return nil, xerror.New("empty response")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, xerror.Wrap(err, "failed to read response body")
	}

	var apiError APIError
	if err := json.Unmarshal(respBody, &apiError); err == nil && apiError.Message != "" {
		return nil, &apiError
	}

	return respBody, nil
}

// SendEmail sends an email
func (c *Client) SendEmail(ctx context.Context, req SendEmailRequest) (*SendEmailResponse, error) {
	if c.debug {
		xlog.Debug("Sending email", "request", req)
	}

	respBody, err := c.sendRequest(ctx, "POST", "/emails", req)
	if err != nil {
		return nil, err
	}

	var sendResp SendEmailResponse
	if err := json.Unmarshal(respBody, &sendResp); err != nil {
		return nil, xerror.Wrap(err, "failed to unmarshal response")
	}

	if c.debug {
		xlog.Debug("Email sent", "response", sendResp)
	}

	return &sendResp, nil
}

// GetEmail retrieves a single email
func (c *Client) GetEmail(ctx context.Context, emailID string) (*Email, error) {
	if c.debug {
		xlog.Debug("Getting email", "emailID", emailID)
	}

	respBody, err := c.sendRequest(ctx, "GET", fmt.Sprintf("/emails/%s", emailID), nil)
	if err != nil {
		return nil, err
	}

	var email Email
	if err := json.Unmarshal(respBody, &email); err != nil {
		return nil, xerror.Wrap(err, "failed to unmarshal response")
	}

	if c.debug {
		xlog.Debug("Got email", "email", email)
	}

	return &email, nil
}

// CreateDomain creates a new domain
func (c *Client) CreateDomain(ctx context.Context, req CreateDomainRequest) (*CreateDomainResponse, error) {
	if c.debug {
		xlog.Debug("Creating domain", "request", req)
	}

	respBody, err := c.sendRequest(ctx, "POST", "/domains", req)
	if err != nil {
		return nil, err
	}

	var resp CreateDomainResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, xerror.Wrap(err, "failed to unmarshal response")
	}

	if c.debug {
		xlog.Debug("Domain created", "response", resp)
	}

	return &resp, nil
}

// ListDomains retrieves a list of domains
func (c *Client) ListDomains(ctx context.Context) (*ListDomainsResponse, error) {
	if c.debug {
		xlog.Debug("Listing domains")
	}

	respBody, err := c.sendRequest(ctx, "GET", "/domains", nil)
	if err != nil {
		return nil, err
	}

	var resp ListDomainsResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, xerror.Wrap(err, "failed to unmarshal response")
	}

	if c.debug {
		xlog.Debug("Domains listed", "response", resp)
	}

	return &resp, nil
}

// GetDomain retrieves a single domain
func (c *Client) GetDomain(ctx context.Context, domainID string) (*Domain, error) {
	if c.debug {
		xlog.Debug("Getting domain", "domainID", domainID)
	}

	respBody, err := c.sendRequest(ctx, "GET", fmt.Sprintf("/domains/%s", domainID), nil)
	if err != nil {
		return nil, err
	}

	var domain Domain
	if err := json.Unmarshal(respBody, &domain); err != nil {
		return nil, xerror.Wrap(err, "failed to unmarshal response")
	}

	if c.debug {
		xlog.Debug("Got domain", "domain", domain)
	}

	return &domain, nil
}

// DeleteDomain removes an existing domain
func (c *Client) DeleteDomain(ctx context.Context, domainID string) error {
	if c.debug {
		xlog.Debug("Deleting domain", "domainID", domainID)
	}

	_, err := c.sendRequest(ctx, "DELETE", fmt.Sprintf("/domains/%s", domainID), nil)
	if err != nil {
		return err
	}

	if c.debug {
		xlog.Debug("Domain deleted", "domainID", domainID)
	}

	return nil
}

// CreateApiKey creates a new API key
func (c *Client) CreateApiKey(ctx context.Context, req CreateApiKeyRequest) (*CreateApiKeyResponse, error) {
	if c.debug {
		xlog.Debug("Creating API key", "request", req)
	}

	respBody, err := c.sendRequest(ctx, "POST", "/api-keys", req)
	if err != nil {
		return nil, err
	}

	var resp CreateApiKeyResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, xerror.Wrap(err, "failed to unmarshal response")
	}

	if c.debug {
		xlog.Debug("API key created", "response", resp)
	}

	return &resp, nil
}

// ListApiKeys retrieves a list of API keys
func (c *Client) ListApiKeys(ctx context.Context) (*ListApiKeysResponse, error) {
	if c.debug {
		xlog.Debug("Listing API keys")
	}

	respBody, err := c.sendRequest(ctx, "GET", "/api-keys", nil)
	if err != nil {
		return nil, err
	}

	var resp ListApiKeysResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, xerror.Wrap(err, "failed to unmarshal response")
	}

	if c.debug {
		xlog.Debug("API keys listed", "response", resp)
	}

	return &resp, nil
}

// DeleteApiKey removes an existing API key
func (c *Client) DeleteApiKey(ctx context.Context, apiKeyID string) error {
	if c.debug {
		xlog.Debug("Deleting API key", "apiKeyID", apiKeyID)
	}

	_, err := c.sendRequest(ctx, "DELETE", fmt.Sprintf("/api-keys/%s", apiKeyID), nil)
	if err != nil {
		return err
	}

	if c.debug {
		xlog.Debug("API key deleted", "apiKeyID", apiKeyID)
	}

	return nil
}

// CreateAudience creates a new audience
func (c *Client) CreateAudience(ctx context.Context, req CreateAudienceRequest) (*CreateAudienceResponse, error) {
	if c.debug {
		xlog.Debug("Creating audience", "request", req)
	}

	var resp CreateAudienceResponse
	body, err := toJSONBody(req)
	if err != nil {
		return nil, err
	}
	err = c.httpClient.PostJSONAndDecode(ctx, "/audiences", body, &resp)
	if err != nil {
		return nil, err
	}

	if c.debug {
		xlog.Debug("Audience created", "response", resp)
	}

	return &resp, nil
}

// ListAudiences retrieves a list of audiences
func (c *Client) ListAudiences(ctx context.Context) (*ListAudiencesResponse, error) {
	if c.debug {
		xlog.Debug("Listing audiences")
	}

	var resp ListAudiencesResponse
	err := c.httpClient.GetJSONAndDecode(ctx, "/audiences", &resp)
	if err != nil {
		return nil, err
	}

	if c.debug {
		xlog.Debug("Audiences listed", "response", resp)
	}

	return &resp, nil
}

// GetAudience retrieves a single audience
func (c *Client) GetAudience(ctx context.Context, audienceID string) (*Audience, error) {
	if c.debug {
		xlog.Debug("Getting audience", "audienceID", audienceID)
	}

	var audience Audience
	err := c.httpClient.GetJSONAndDecode(ctx, fmt.Sprintf("/audiences/%s", audienceID), &audience)
	if err != nil {
		return nil, err
	}

	if c.debug {
		xlog.Debug("Got audience", "audience", audience)
	}

	return &audience, nil
}

// DeleteAudience removes an existing audience
func (c *Client) DeleteAudience(ctx context.Context, audienceID string) error {
	if c.debug {
		xlog.Debug("Deleting audience", "audienceID", audienceID)
	}

	_, err := c.httpClient.Delete(ctx, fmt.Sprintf("/audiences/%s", audienceID))
	if err != nil {
		return handleAPIError(err)
	}

	if c.debug {
		xlog.Debug("Audience deleted", "audienceID", audienceID)
	}

	return nil
}

// CreateContact creates a new contact in an audience
func (c *Client) CreateContact(ctx context.Context, audienceID string, req CreateContactRequest) (*CreateContactResponse, error) {
	if audienceID == "" {
		return nil, xerror.New("audienceID cannot be empty")
	}

	if c.debug {
		xlog.Debug("Creating contact", "audienceID", audienceID, "request", req)
	}

	var resp CreateContactResponse
	body, err := toJSONBody(req)
	if err != nil {
		return nil, xerror.Wrap(err, "failed to prepare request body")
	}

	err = c.httpClient.PostJSONAndDecode(ctx, fmt.Sprintf("/audiences/%s/contacts", audienceID), body, &resp)
	if err != nil {
		return nil, xerror.Wrap(err, "failed to create contact")
	}

	if c.debug {
		xlog.Debug("Contact created", "response", resp)
	}

	return &resp, nil
}

// ListContacts retrieves a list of contacts in an audience
func (c *Client) ListContacts(ctx context.Context, audienceID string) (*ListContactsResponse, error) {
	if c.debug {
		xlog.Debug("Listing contacts", "audienceID", audienceID)
	}

	var resp ListContactsResponse
	err := c.httpClient.GetJSONAndDecode(ctx, fmt.Sprintf("/audiences/%s/contacts", audienceID), &resp)
	if err != nil {
		return nil, err
	}

	if c.debug {
		xlog.Debug("Contacts listed", "response", resp)
	}

	return &resp, nil
}

// GetContact retrieves a single contact
func (c *Client) GetContact(ctx context.Context, audienceID, contactID string) (*Contact, error) {
	if c.debug {
		xlog.Debug("Getting contact", "audienceID", audienceID, "contactID", contactID)
	}

	var contact Contact
	err := c.httpClient.GetJSONAndDecode(ctx, fmt.Sprintf("/audiences/%s/contacts/%s", audienceID, contactID), &contact)
	if err != nil {
		return nil, err
	}

	if c.debug {
		xlog.Debug("Got contact", "contact", contact)
	}

	return &contact, nil
}

// DeleteContact removes an existing contact
func (c *Client) DeleteContact(ctx context.Context, audienceID, contactID string) error {
	if c.debug {
		xlog.Debug("Deleting contact", "audienceID", audienceID, "contactID", contactID)
	}

	_, err := c.httpClient.Delete(ctx, fmt.Sprintf("/audiences/%s/contacts/%s", audienceID, contactID))
	if err != nil {
		return handleAPIError(err)
	}

	if c.debug {
		xlog.Debug("Contact deleted", "audienceID", audienceID, "contactID", contactID)
	}

	return nil
}
