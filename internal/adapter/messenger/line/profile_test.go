package line

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_GetProfile_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/bot/profile/U1234567890", r.URL.Path)
		assert.Equal(t, "Bearer test_token", r.Header.Get("Authorization"))
		json.NewEncoder(w).Encode(LineProfile{
			DisplayName: "Test User",
			Language:    "zh-TW",
		})
	}))
	defer server.Close()

	client := &Client{
		channelToken: "test_token",
		apiURL:       server.URL + "/v2/bot/message",
		profileURL:   server.URL + "/v2/bot/profile",
		httpClient:   server.Client(),
	}

	profile, err := client.GetProfile(context.Background(), "U1234567890")
	assert.NoError(t, err)
	assert.Equal(t, "Test User", profile.DisplayName)
	assert.Equal(t, "zh-TW", profile.Language)
}

func TestClient_GetProfile_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "user not found"})
	}))
	defer server.Close()

	client := &Client{
		channelToken: "test_token",
		apiURL:       server.URL + "/v2/bot/message",
		profileURL:   server.URL + "/v2/bot/profile",
		httpClient:   server.Client(),
	}

	_, err := client.GetProfile(context.Background(), "U_INVALID")
	assert.Error(t, err)
}
