# XResend - Go Client for Resend API

XResend is a powerful and flexible Go client for interacting with the Resend API. It provides a simple and intuitive interface for managing emails, domains, API keys, audiences, and contacts.

## Features

- Easy-to-use client for Resend API
- Support for sending emails with attachments and tags
- Domain management (create, list, get, delete)
- API key management (create, list, delete)
- Audience management (create, list, get, delete)
- Contact management (create, list, get, delete)
- Customizable HTTP client with debugging options
- Error handling with custom API error type

## Installation

To install XResend, use `go get`:

```bash
go get github.com/seefs001/xox/xresend
```

## Usage

### Creating a Client

```go
import "github.com/seefs001/xox/xresend"

// Create a new client
client, err := xresend.NewClient("your_api_key")
if err != nil {
    // Handle error
}

// Create a client with debug mode enabled
client, err := xresend.NewClient("your_api_key", xresend.WithDebug(true))
if err != nil {
    // Handle error
}
```

### Sending an Email

```go
req := xresend.SendEmailRequest{
    From:    "sender@example.com",
    To:      []string{"recipient@example.com"},
    Subject: "Hello from XResend",
    HTML:    "<p>This is a test email sent using XResend.</p>",
}

resp, err := client.SendEmail(context.Background(), req)
if err != nil {
    // Handle error
}
fmt.Printf("Email sent with ID: %s\n", resp.ID)
```

### Managing Domains

```go
// Create a domain
createReq := xresend.CreateDomainRequest{
    Name:   "example.com",
    Region: "us-east-1",
}
domain, err := client.CreateDomain(context.Background(), createReq)
if err != nil {
    // Handle error
}

// List domains
domains, err := client.ListDomains(context.Background())
if err != nil {
    // Handle error
}

// Get a domain
domain, err := client.GetDomain(context.Background(), "domain_id")
if err != nil {
    // Handle error
}

// Delete a domain
err := client.DeleteDomain(context.Background(), "domain_id")
if err != nil {
    // Handle error
}
```

### Managing API Keys

```go
// Create an API key
createReq := xresend.CreateApiKeyRequest{
    Name:       "My API Key",
    Permission: "full_access",
}
apiKey, err := client.CreateApiKey(context.Background(), createReq)
if err != nil {
    // Handle error
}

// List API keys
apiKeys, err := client.ListApiKeys(context.Background())
if err != nil {
    // Handle error
}

// Delete an API key
err := client.DeleteApiKey(context.Background(), "api_key_id")
if err != nil {
    // Handle error
}
```

### Managing Audiences and Contacts

```go
// Create an audience
audienceReq := xresend.CreateAudienceRequest{
    Name: "My Audience",
}
audience, err := client.CreateAudience(context.Background(), audienceReq)
if err != nil {
    // Handle error
}

// Create a contact
contactReq := xresend.CreateContactRequest{
    Email:     "contact@example.com",
    FirstName: "John",
    LastName:  "Doe",
}
contact, err := client.CreateContact(context.Background(), audience.ID, contactReq)
if err != nil {
    // Handle error
}

// List contacts
contacts, err := client.ListContacts(context.Background(), audience.ID)
if err != nil {
    // Handle error
}
```

## Error Handling

XResend uses a custom `APIError` type for API-specific errors. You can check for these errors and handle them accordingly:

```go
_, err := client.SendEmail(context.Background(), req)
if err != nil {
    if apiErr, ok := err.(*xresend.APIError); ok {
        fmt.Printf("API Error: %s (Status: %d, Name: %s)\n", apiErr.Message, apiErr.StatusCode, apiErr.Name)
    } else {
        fmt.Printf("Error: %v\n", err)
    }
}
```

## Debugging

You can enable debug mode to log detailed information about requests and responses:

```go
client, err := xresend.NewClient("your_api_key", xresend.WithDebug(true))
```

## Best Practices

1. Always check for errors returned by the client methods.
2. Use context for better control over request cancellation and timeouts.
3. Enable debug mode during development to troubleshoot API interactions.
4. Securely manage your API keys and never hardcode them in your source code.
5. Use environment variables or a secure configuration management system to store API keys.

## Contributing

Contributions to XResend are welcome! Please feel free to submit issues, fork the repository and send pull requests!

## License

XResend is released under the MIT License. See the LICENSE file for details.

