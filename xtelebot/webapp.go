package xtelebot

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

var (
	ErrInvalidHash     = errors.New("invalid hash")
	ErrMissingHash     = errors.New("missing hash")
	ErrMissingAuthDate = errors.New("missing auth_date")
	ErrDataExpired     = errors.New("init data expired")
)

// ValidateWebAppData validates the web app data received from the Telegram Mini App
func ValidateWebAppData(initData, botToken string, maxAge time.Duration) error {
	params, err := url.ParseQuery(initData)
	if err != nil {
		return err
	}

	hash := params.Get("hash")
	if hash == "" {
		return ErrMissingHash
	}
	params.Del("hash")

	authDateStr := params.Get("auth_date")
	if authDateStr == "" {
		return ErrMissingAuthDate
	}

	authDate, err := time.Parse(time.RFC3339, authDateStr)
	if err != nil {
		return err
	}

	if maxAge > 0 && time.Since(authDate) > maxAge {
		return ErrDataExpired
	}

	dataCheckString := buildDataCheckString(params)
	secretKey := generateSecretKey(botToken)

	if !verifySignature(dataCheckString, secretKey, hash) {
		return ErrInvalidHash
	}

	return nil
}

func buildDataCheckString(params url.Values) string {
	pairs := make([]string, 0, len(params))
	for key, values := range params {
		pairs = append(pairs, key+"="+values[0])
	}
	sort.Strings(pairs)
	return strings.Join(pairs, "\n")
}

func generateSecretKey(botToken string) []byte {
	h := hmac.New(sha256.New, []byte("WebAppData"))
	h.Write([]byte(botToken))
	return h.Sum(nil)
}

func verifySignature(data string, secretKey []byte, expectedHash string) bool {
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(data))
	signature := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedHash))
}

// InitData contains init data.
// https://core.telegram.org/bots/webapps#webappinitdata
type InitData struct {
	QueryID      string     `json:"query_id"`
	User         WebAppUser `json:"user"`
	Receiver     WebAppUser `json:"receiver"`
	Chat         WebAppChat `json:"chat"`
	ChatType     string     `json:"chat_type"`
	ChatInstance string     `json:"chat_instance"`
	StartParam   string     `json:"start_param"`
	CanSendAfter int        `json:"can_send_after"`
	AuthDate     int        `json:"auth_date"`
	Hash         string     `json:"hash"`
}

// WebAppUser represents user data in the Mini App init data.
type WebAppUser struct {
	ID                    int64  `json:"id"`
	IsBot                 bool   `json:"is_bot,omitempty"`
	FirstName             string `json:"first_name"`
	LastName              string `json:"last_name,omitempty"`
	Username              string `json:"username,omitempty"`
	LanguageCode          string `json:"language_code,omitempty"`
	IsPremium             bool   `json:"is_premium,omitempty"`
	AddedToAttachmentMenu bool   `json:"added_to_attachment_menu,omitempty"`
	AllowsWriteToPM       bool   `json:"allows_write_to_pm,omitempty"`
	PhotoURL              string `json:"photo_url,omitempty"`
}

// WebAppChat represents chat data in the Mini App init data.
type WebAppChat struct {
	ID       int64  `json:"id"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	Username string `json:"username,omitempty"`
	PhotoURL string `json:"photo_url,omitempty"`
}

func Parse(initData string) (InitData, error) {
	params, err := url.ParseQuery(initData)
	if err != nil {
		return InitData{}, fmt.Errorf("invalid init data format: %w", err)
	}

	jsonObj := make(map[string]interface{})

	for key, values := range params {
		if len(values) == 0 {
			continue
		}

		value := values[0]
		jsonObj[key] = value
	}

	jsonData, err := json.Marshal(jsonObj)
	if err != nil {
		return InitData{}, fmt.Errorf("failed to marshal data: %w", err)
	}

	var initDataObj InitData
	if err := json.Unmarshal(jsonData, &initDataObj); err != nil {
		return InitData{}, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return initDataObj, nil
}
