package xtelebot_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/seefs001/xox/xtelebot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBot(t *testing.T) {
	t.Run("Valid token", func(t *testing.T) {
		bot, err := xtelebot.NewBot("valid_token")
		assert.NoError(t, err)
		assert.NotNil(t, bot)
	})

	t.Run("Empty token", func(t *testing.T) {
		bot, err := xtelebot.NewBot("")
		assert.Error(t, err)
		assert.Nil(t, bot)
	})
}

func TestBotOptions(t *testing.T) {
	t.Run("WithBaseURL", func(t *testing.T) {
		bot, err := xtelebot.NewBot("token", xtelebot.WithBaseURL("https://custom.api.telegram.org"))
		require.NoError(t, err)
		assert.NotNil(t, bot)
		// Note: We can't directly test the baseURL as it's private, but we can test its effect in other methods
	})

	t.Run("WithDebug", func(t *testing.T) {
		bot, err := xtelebot.NewBot("token", xtelebot.WithDebug(true))
		require.NoError(t, err)
		assert.NotNil(t, bot)
		// Note: Debug mode is also private, but its effects can be tested in logging output
	})
}

func TestGetMe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/bottoken/getMe", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true,"result":{"id":123,"is_bot":true,"first_name":"TestBot","username":"test_bot"}}`))
	}))
	defer server.Close()

	bot, err := xtelebot.NewBot("token", xtelebot.WithBaseURL(server.URL))
	require.NoError(t, err)

	user, err := bot.GetMe(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, 123, user.ID)
	assert.True(t, user.IsBot)
	assert.Equal(t, "TestBot", user.FirstName)
	assert.Equal(t, "test_bot", user.Username)
}

func TestSendMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/bottoken/sendMessage", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		err := r.ParseMultipartForm(10 << 20) // 10 MB max memory
		assert.NoError(t, err)
		assert.Equal(t, "123", r.FormValue("chat_id"))
		assert.Equal(t, "Test message", r.FormValue("text"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true,"result":{"message_id":1,"text":"Test message"}}`))
	}))
	defer server.Close()

	bot, err := xtelebot.NewBot("token", xtelebot.WithBaseURL(server.URL))
	require.NoError(t, err)

	msg, err := bot.SendMessage(context.Background(), int64(123), "Test message")
	assert.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, 1, msg.MessageID)
	assert.Equal(t, "Test message", msg.Text)
}

func TestGetUpdates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/bottoken/getUpdates", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		r.ParseForm()
		assert.Equal(t, "10", r.Form.Get("offset"))
		assert.Equal(t, "5", r.Form.Get("limit"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true,"result":[{"update_id":1,"message":{"message_id":1,"text":"Test update"}}]}`))
	}))
	defer server.Close()

	bot, err := xtelebot.NewBot("token", xtelebot.WithBaseURL(server.URL))
	require.NoError(t, err)

	updates, err := bot.GetUpdates(context.Background(), 10, 5)
	assert.NoError(t, err)
	assert.Len(t, updates, 1)
	assert.Equal(t, 1, updates[0].UpdateID)
	assert.Equal(t, "Test update", updates[0].Message.Text)
}

func TestSendPhoto(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/bottoken/sendPhoto", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		err := r.ParseMultipartForm(10 << 20) // 10 MB max memory
		assert.NoError(t, err)
		assert.Equal(t, "123", r.FormValue("chat_id"))
		assert.Equal(t, "photo_file_id", r.FormValue("photo"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true,"result":{"message_id":1,"photo":[{"file_id":"photo_file_id"}]}}`))
	}))
	defer server.Close()

	bot, err := xtelebot.NewBot("token", xtelebot.WithBaseURL(server.URL))
	require.NoError(t, err)

	msg, err := bot.SendPhoto(context.Background(), int64(123), "photo_file_id")
	assert.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, 1, msg.MessageID)
	assert.Len(t, msg.Photo, 1)
	assert.Equal(t, "photo_file_id", msg.Photo[0].FileID)
}

func TestAnswerCallbackQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/bottoken/answerCallbackQuery", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		err := r.ParseMultipartForm(10 << 20) // 10 MB max memory
		assert.NoError(t, err)
		assert.Equal(t, "query_id", r.FormValue("callback_query_id"))
		assert.Equal(t, "Test answer", r.FormValue("text"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true,"result":true}`))
	}))
	defer server.Close()

	bot, err := xtelebot.NewBot("token", xtelebot.WithBaseURL(server.URL))
	require.NoError(t, err)

	err = bot.AnswerCallbackQuery(context.Background(), "query_id", xtelebot.WithCallbackQueryText("Test answer"))
	assert.NoError(t, err)
}
