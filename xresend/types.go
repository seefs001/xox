package xresend

import (
	"fmt"
	"strings"
	"time"
)

// SendEmailRequest represents the request body for sending an email
type SendEmailRequest struct {
	From        string       `json:"from"`
	To          []string     `json:"to"`
	Subject     string       `json:"subject"`
	Bcc         []string     `json:"bcc,omitempty"`
	Cc          []string     `json:"cc,omitempty"`
	ReplyTo     []string     `json:"reply_to,omitempty"`
	HTML        string       `json:"html,omitempty"`
	Text        string       `json:"text,omitempty"`
	Headers     interface{}  `json:"headers,omitempty"`
	ScheduledAt string       `json:"scheduled_at,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	Tags        []Tag        `json:"tags,omitempty"`
}

// AddRecipient adds a recipient to the To field
func (r *SendEmailRequest) AddRecipient(email string) {
	r.To = append(r.To, email)
}

// AddCC adds a recipient to the Cc field
func (r *SendEmailRequest) AddCC(email string) {
	r.Cc = append(r.Cc, email)
}

// AddBCC adds a recipient to the Bcc field
func (r *SendEmailRequest) AddBCC(email string) {
	r.Bcc = append(r.Bcc, email)
}

// AddAttachment adds an attachment to the email
func (r *SendEmailRequest) AddAttachment(attachment Attachment) {
	r.Attachments = append(r.Attachments, attachment)
}

// AddTag adds a tag to the email
func (r *SendEmailRequest) AddTag(name, value string) {
	r.Tags = append(r.Tags, Tag{Name: name, Value: value})
}

// Attachment represents an email attachment
type Attachment struct {
	Content     string `json:"content"`
	Filename    string `json:"filename"`
	Path        string `json:"path,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

// Tag represents an email tag
type Tag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// SendEmailResponse represents the response from sending an email
type SendEmailResponse struct {
	ID string `json:"id"`
}

// Email represents a detailed email object
type Email struct {
	Object    string     `json:"object"`
	ID        string     `json:"id"`
	To        []string   `json:"to"`
	From      string     `json:"from"`
	CreatedAt CustomTime `json:"created_at"`
	Subject   string     `json:"subject"`
	HTML      string     `json:"html"`
	Text      string     `json:"text"`
	Bcc       []string   `json:"bcc"`
	Cc        []string   `json:"cc"`
	ReplyTo   []string   `json:"reply_to"`
	LastEvent string     `json:"last_event"`
}

// Recipients returns all recipients (To, Cc, Bcc) of the email
func (e *Email) Recipients() []string {
	recipients := make([]string, 0, len(e.To)+len(e.Cc)+len(e.Bcc))
	recipients = append(recipients, e.To...)
	recipients = append(recipients, e.Cc...)
	recipients = append(recipients, e.Bcc...)
	return recipients
}

// IsDelivered checks if the last event indicates the email was delivered
func (e *Email) IsDelivered() bool {
	return e.LastEvent == "delivered"
}

// CreateDomainRequest represents the request body for creating a domain
type CreateDomainRequest struct {
	Name   string `json:"name"`
	Region string `json:"region,omitempty"`
}

// CreateDomainResponse represents the response from creating a domain
type CreateDomainResponse struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	Status    string         `json:"status"`
	Records   []DomainRecord `json:"records"`
	Region    string         `json:"region"`
}

// DomainRecord represents a DNS record for a domain
type DomainRecord struct {
	Record   string `json:"record"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	TTL      string `json:"ttl"`
	Status   string `json:"status"`
	Value    string `json:"value"`
	Priority int    `json:"priority,omitempty"`
}

// Domain represents a detailed domain object
type Domain struct {
	Object    string         `json:"object"`
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Status    string         `json:"status"`
	CreatedAt CustomTime     `json:"created_at"`
	Region    string         `json:"region"`
	Records   []DomainRecord `json:"records"`
}

// ListDomainsResponse represents the response from listing domains
type ListDomainsResponse struct {
	Data []Domain `json:"data"`
}

// CreateApiKeyRequest represents the request body for creating an API key
type CreateApiKeyRequest struct {
	Name       string `json:"name"`
	Permission string `json:"permission,omitempty"`
	DomainID   string `json:"domain_id,omitempty"`
}

// CreateApiKeyResponse represents the response from creating an API key
type CreateApiKeyResponse struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

// ListApiKeysResponse represents the response from listing API keys
type ListApiKeysResponse struct {
	Data []ApiKey `json:"data"`
}

// ApiKey represents an API key object
type ApiKey struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateAudienceRequest represents the request body for creating an audience
type CreateAudienceRequest struct {
	Name string `json:"name"`
}

// CreateAudienceResponse represents the response from creating an audience
type CreateAudienceResponse struct {
	ID     string `json:"id"`
	Object string `json:"object"`
	Name   string `json:"name"`
}

// Audience represents an audience object
type Audience struct {
	ID        string    `json:"id"`
	Object    string    `json:"object"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// ListAudiencesResponse represents the response from listing audiences
type ListAudiencesResponse struct {
	Object string     `json:"object"`
	Data   []Audience `json:"data"`
}

// CreateContactRequest represents the request body for creating a contact
type CreateContactRequest struct {
	Email        string `json:"email"`
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	Unsubscribed bool   `json:"unsubscribed,omitempty"`
}

// CreateContactResponse represents the response from creating a contact
type CreateContactResponse struct {
	Object string `json:"object"`
	ID     string `json:"id"`
}

// Contact represents a contact object
type Contact struct {
	Object       string    `json:"object"`
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	CreatedAt    time.Time `json:"created_at"`
	Unsubscribed bool      `json:"unsubscribed"`
}

// ListContactsResponse represents the response from listing contacts
type ListContactsResponse struct {
	Object string    `json:"object"`
	Data   []Contact `json:"data"`
}

// CustomTime is a custom time type that can parse various time formats returned by Resend API
type CustomTime struct {
	time.Time
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")

	formats := []string{
		"2006-01-02 15:04:05.999999-07:00",
		"2006-01-02 15:04:05.999999+00",
		"2006-01-02T15:04:05.999999Z07:00",
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			ct.Time = t
			return nil
		}
	}

	return fmt.Errorf("unable to parse time: %s", s)
}
