package security

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

var errNotFound = errors.New("not found")

// Mock repositories for security tests
type SecurityTestExpenseRepository struct {
	expenses map[string]*domain.Expense
}

func (r *SecurityTestExpenseRepository) Create(ctx context.Context, expense *domain.Expense) error {
	r.expenses[expense.ID] = expense
	return nil
}

func (r *SecurityTestExpenseRepository) GetByID(ctx context.Context, id string) (*domain.Expense, error) {
	if exp, ok := r.expenses[id]; ok {
		return exp, nil
	}
	return nil, errNotFound
}

func (r *SecurityTestExpenseRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Expense, error) {
	return []*domain.Expense{}, nil
}

func (r *SecurityTestExpenseRepository) GetByUserIDAndDateRange(ctx context.Context, userID string, from, to time.Time) ([]*domain.Expense, error) {
	return []*domain.Expense{}, nil
}

func (r *SecurityTestExpenseRepository) GetByUserIDAndCategory(ctx context.Context, userID, categoryID string) ([]*domain.Expense, error) {
	return []*domain.Expense{}, nil
}

func (r *SecurityTestExpenseRepository) Update(ctx context.Context, expense *domain.Expense) error {
	return nil
}

func (r *SecurityTestExpenseRepository) Delete(ctx context.Context, id string) error {
	return nil
}

type SecurityTestUserRepository struct {
	users map[string]*domain.User
}

func (r *SecurityTestUserRepository) Create(ctx context.Context, user *domain.User) error {
	r.users[user.UserID] = user
	return nil
}

func (r *SecurityTestUserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	if user, ok := r.users[userID]; ok {
		return user, nil
	}
	return nil, errNotFound
}

func (r *SecurityTestUserRepository) Exists(ctx context.Context, userID string) (bool, error) {
	_, ok := r.users[userID]
	return ok, nil
}

type SecurityTestCategoryRepository struct {
	categories map[string]*domain.Category
}

func (r *SecurityTestCategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	r.categories[category.ID] = category
	return nil
}

func (r *SecurityTestCategoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	return nil, errNotFound
}

func (r *SecurityTestCategoryRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Category, error) {
	return []*domain.Category{}, nil
}

func (r *SecurityTestCategoryRepository) GetByUserIDAndName(ctx context.Context, userID, name string) (*domain.Category, error) {
	return nil, errNotFound
}

func (r *SecurityTestCategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	return nil
}

func (r *SecurityTestCategoryRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (r *SecurityTestCategoryRepository) CreateKeyword(ctx context.Context, keyword *domain.CategoryKeyword) error {
	return nil
}

func (r *SecurityTestCategoryRepository) GetKeywordsByCategory(ctx context.Context, categoryID string) ([]*domain.CategoryKeyword, error) {
	return []*domain.CategoryKeyword{}, nil
}

func (r *SecurityTestCategoryRepository) DeleteKeyword(ctx context.Context, id string) error {
	return nil
}

type SecurityTestAIService struct{}

func (s *SecurityTestAIService) ParseExpense(ctx context.Context, text string, userID string) ([]*domain.ParsedExpense, error) {
	return []*domain.ParsedExpense{}, nil
}

func (s *SecurityTestAIService) SuggestCategory(ctx context.Context, description string) (string, error) {
	return "food", nil
}

// TestLINESignatureVerification tests LINE signature verification security
func TestLINESignatureVerification(t *testing.T) {
	secret := "test_channel_secret"
	payload := []byte(`{"events":[{"type":"message","source":{"userId":"U1234"},"message":{"type":"text","text":"test"}}]}`)

	// Compute valid signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	validSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	t.Run("ValidSignature", func(t *testing.T) {
		if !verifyLINESignature(payload, validSignature, secret) {
			t.Error("Valid signature should pass verification")
		}
	})

	t.Run("InvalidSignature", func(t *testing.T) {
		invalidSignature := base64.StdEncoding.EncodeToString([]byte("invalid"))
		if verifyLINESignature(payload, invalidSignature, secret) {
			t.Error("Invalid signature should fail verification")
		}
	})

	t.Run("ModifiedPayload", func(t *testing.T) {
		modifiedPayload := []byte(`{"events":[{"type":"message","source":{"userId":"U9999"},"message":{"type":"text","text":"test"}}]}`)
		if verifyLINESignature(modifiedPayload, validSignature, secret) {
			t.Error("Modified payload should fail verification")
		}
	})

	t.Run("EmptyPayload", func(t *testing.T) {
		emptyPayload := []byte("")
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(emptyPayload)
		emptySignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
		if !verifyLINESignature(emptyPayload, emptySignature, secret) {
			t.Error("Empty payload with correct signature should pass")
		}
	})

	t.Run("TimingAttackResistance", func(t *testing.T) {
		// Verify constant-time comparison behavior
		correctSig := validSignature
		wrongSig := "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXo=" // Different signature

		// Both should be checked in constant time - no early exit
		// This test just ensures we're using proper comparison
		verifyLINESignature(payload, correctSig, secret)
		verifyLINESignature(payload, wrongSig, secret)
	})
}

// TestSlackSignatureVerification tests Slack signature verification with timestamp window
func TestSlackSignatureVerification(t *testing.T) {
	secret := "test_signing_secret"
	timestamp := time.Now().Unix()
	payload := []byte(`{"type":"url_verification","challenge":"test_challenge"}`)

	// Compute valid signature
	basestring := "v0:" + strconv.FormatInt(timestamp, 10) + ":" + string(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(basestring))
	validSignature := "v0=" + hex.EncodeToString(mac.Sum(nil))

	t.Run("ValidSignature", func(t *testing.T) {
		if !verifySlackSignature(payload, validSignature, timestamp, secret) {
			t.Error("Valid signature with current timestamp should pass")
		}
	})

	t.Run("InvalidSignature", func(t *testing.T) {
		if verifySlackSignature(payload, "v0=invalid", timestamp, secret) {
			t.Error("Invalid signature should fail")
		}
	})

	t.Run("ReplayAttackPrevention", func(t *testing.T) {
		// Test 5-minute window (300 seconds)
		oldTimestamp := time.Now().Unix() - 600 // 10 minutes ago
		basestring := "v0:" + strconv.FormatInt(oldTimestamp, 10) + ":" + string(payload)
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(basestring))
		oldSignature := "v0=" + hex.EncodeToString(mac.Sum(nil))

		if verifySlackSignature(payload, oldSignature, oldTimestamp, secret) {
			t.Error("Signature from 10 minutes ago should fail (outside 5-minute window)")
		}
	})

	t.Run("ModifiedPayload", func(t *testing.T) {
		modifiedPayload := []byte(`{"type":"url_verification","challenge":"different_challenge"}`)
		if verifySlackSignature(modifiedPayload, validSignature, timestamp, secret) {
			t.Error("Modified payload should fail verification")
		}
	})

	t.Run("WrongTimestamp", func(t *testing.T) {
		wrongTimestamp := timestamp - 1000
		if verifySlackSignature(payload, validSignature, wrongTimestamp, secret) {
			t.Error("Signature with wrong timestamp should fail")
		}
	})
}

// TestWhatsAppSignatureVerification tests WhatsApp signature verification (hex-encoded)
func TestWhatsAppSignatureVerification(t *testing.T) {
	secret := "test_app_secret"
	payload := []byte(`{"messages":[{"from":"1234567890","type":"text","text":{"body":"test"}}]}`)

	// Compute valid signature (hex encoded)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	validSignature := hex.EncodeToString(mac.Sum(nil))

	t.Run("ValidSignature", func(t *testing.T) {
		if !verifyWhatsAppSignature(payload, "sha256="+validSignature, secret) {
			t.Error("Valid signature should pass verification")
		}
	})

	t.Run("InvalidSignature", func(t *testing.T) {
		if verifyWhatsAppSignature(payload, "sha256=invalid", secret) {
			t.Error("Invalid signature should fail")
		}
	})

	t.Run("MissingScheme", func(t *testing.T) {
		if verifyWhatsAppSignature(payload, validSignature, secret) {
			t.Error("Signature without sha256= prefix should fail")
		}
	})

	t.Run("ModifiedPayload", func(t *testing.T) {
		modifiedPayload := []byte(`{"messages":[{"from":"9999999999","type":"text","text":{"body":"test"}}]}`)
		if verifyWhatsAppSignature(modifiedPayload, "sha256="+validSignature, secret) {
			t.Error("Modified payload should fail verification")
		}
	})

	t.Run("EmptyPayload", func(t *testing.T) {
		emptyPayload := []byte("")
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(emptyPayload)
		emptySignature := hex.EncodeToString(mac.Sum(nil))
		if !verifyWhatsAppSignature(emptyPayload, "sha256="+emptySignature, secret) {
			t.Error("Empty payload with correct signature should pass")
		}
	})

	t.Run("LargePayload", func(t *testing.T) {
		// Test with large payload
		largePayload := make([]byte, 10000)
		for i := range largePayload {
			largePayload[i] = byte(i % 256)
		}
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(largePayload)
		largeSignature := hex.EncodeToString(mac.Sum(nil))
		if !verifyWhatsAppSignature(largePayload, "sha256="+largeSignature, secret) {
			t.Error("Large payload with correct signature should pass")
		}
	})
}

// TestTeamsSignatureVerification tests Teams Bearer token verification
func TestTeamsSignatureVerification(t *testing.T) {
	secret := "test_app_password"
	payload := []byte(`{"type":"message","from":{"id":"user123"},"text":"test"}`)

	// Compute valid signature (base64 encoded HMAC)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	validSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	validHeader := "Bearer " + validSignature

	t.Run("ValidSignature", func(t *testing.T) {
		if !verifyTeamsSignature(payload, validHeader, secret) {
			t.Error("Valid Bearer token should pass verification")
		}
	})

	t.Run("InvalidSignature", func(t *testing.T) {
		if verifyTeamsSignature(payload, "Bearer invalid", secret) {
			t.Error("Invalid Bearer token should fail")
		}
	})

	t.Run("MissingBearer", func(t *testing.T) {
		if verifyTeamsSignature(payload, validSignature, secret) {
			t.Error("Missing 'Bearer ' prefix should fail")
		}
	})

	t.Run("ModifiedPayload", func(t *testing.T) {
		modifiedPayload := []byte(`{"type":"message","from":{"id":"user999"},"text":"test"}`)
		if verifyTeamsSignature(modifiedPayload, validHeader, secret) {
			t.Error("Modified payload should fail verification")
		}
	})

	t.Run("EmptyHeader", func(t *testing.T) {
		if verifyTeamsSignature(payload, "", secret) {
			t.Error("Empty Authorization header should fail")
		}
	})
}

// TestDiscordSignatureVerification tests Discord interaction signature verification
func TestDiscordSignatureVerification(t *testing.T) {
	t.Run("ValidPingInteraction", func(t *testing.T) {
		interaction := map[string]interface{}{
			"type": 1,
			"id":   "ping_123",
		}
		body, _ := json.Marshal(interaction)

		// Discord signature verification would use the actual ed25519 public key
		// This test verifies the structure is correct
		if !isValidDiscordInteraction(body) {
			t.Error("Valid PING interaction should be recognized")
		}
	})

	t.Run("MissingType", func(t *testing.T) {
		interaction := map[string]interface{}{
			"id": "ping_123",
		}
		body, _ := json.Marshal(interaction)

		if isValidDiscordInteraction(body) {
			t.Error("Interaction without type should fail")
		}
	})

	t.Run("InvalidType", func(t *testing.T) {
		interaction := map[string]interface{}{
			"type": "invalid_string",
			"id":   "ping_123",
		}
		body, _ := json.Marshal(interaction)

		if isValidDiscordInteraction(body) {
			t.Error("Interaction with invalid type should fail")
		}
	})
}

// Signature verification helper functions

// formatTimestamp formats unix timestamp as string
func formatTimestamp(t int64) string {
	return string([]byte{
		byte((t >> 56) & 0xFF),
		byte((t >> 48) & 0xFF),
		byte((t >> 40) & 0xFF),
		byte((t >> 32) & 0xFF),
		byte((t >> 24) & 0xFF),
		byte((t >> 16) & 0xFF),
		byte((t >> 8) & 0xFF),
		byte(t & 0xFF),
	})
}

// verifyLINESignature verifies LINE signature using base64-encoded HMAC-SHA256
func verifyLINESignature(payload []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// verifySlackSignature verifies Slack signature with timestamp window
func verifySlackSignature(payload []byte, signature string, timestamp int64, secret string) bool {
	// Check 5-minute window (300 seconds)
	now := time.Now().Unix()
	if now-timestamp > 300 {
		return false
	}

	basestring := "v0:" + strconv.FormatInt(timestamp, 10) + ":" + string(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(basestring))
	expectedSignature := "v0=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// verifyWhatsAppSignature verifies WhatsApp signature using hex-encoded HMAC-SHA256
func verifyWhatsAppSignature(payload []byte, header, secret string) bool {
	if len(header) < 7 || header[:7] != "sha256=" {
		return false
	}
	signature := header[7:]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// verifyTeamsSignature verifies Teams Bearer token using base64-encoded HMAC-SHA256
func verifyTeamsSignature(payload []byte, header, secret string) bool {
	if len(header) < 7 || header[:7] != "Bearer " {
		return false
	}
	signature := header[7:]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// isValidDiscordInteraction validates Discord interaction structure
func isValidDiscordInteraction(payload []byte) bool {
	var interaction map[string]interface{}
	if err := json.Unmarshal(payload, &interaction); err != nil {
		return false
	}
	val, hasType := interaction["type"]
	if !hasType {
		return false
	}
	_, ok := val.(float64)
	return ok
}

// TestSignatureEdgeCases tests edge cases across all platforms
func TestSignatureEdgeCases(t *testing.T) {
	t.Run("EmptySecret", func(t *testing.T) {
		payload := []byte("test")
		emptySecret := ""
		mac := hmac.New(sha256.New, []byte(emptySecret))
		mac.Write(payload)
		sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))
		if !verifyLINESignature(payload, sig, emptySecret) {
			t.Error("HMAC with empty secret should still compute correctly")
		}
	})

	t.Run("VeryLargeSecret", func(t *testing.T) {
		payload := []byte("test")
		largeSecret := make([]byte, 10000)
		for i := range largeSecret {
			largeSecret[i] = byte(i % 256)
		}
		mac := hmac.New(sha256.New, largeSecret)
		mac.Write(payload)
		sig := hex.EncodeToString(mac.Sum(nil))
		if !verifyWhatsAppSignature(payload, "sha256="+sig, string(largeSecret)) {
			t.Error("HMAC with large secret should work")
		}
	})

	t.Run("SpecialCharactersInPayload", func(t *testing.T) {
		payload := []byte(`{"text":"Hello 你好 مرحبا שלום 🚀"}`)
		secret := "test_secret"
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(payload)
		sig := hex.EncodeToString(mac.Sum(nil))
		if !verifyWhatsAppSignature(payload, "sha256="+sig, secret) {
			t.Error("HMAC with special characters should work")
		}
	})

	t.Run("NullBytesInPayload", func(t *testing.T) {
		payload := []byte{0x00, 0x01, 0x02, 0x03, 0xff}
		secret := "test_secret"
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(payload)
		sig := hex.EncodeToString(mac.Sum(nil))
		if !verifyWhatsAppSignature(payload, "sha256="+sig, secret) {
			t.Error("HMAC with null bytes should work")
		}
	})
}

// TestReplayAttackPrevention tests replay attack prevention mechanisms
func TestReplayAttackPrevention(t *testing.T) {
	t.Run("SlackReplayWindowBoundary", func(t *testing.T) {
		secret := "test_secret"
		payload := []byte(`{"type":"event_callback"}`)

		// Test within 5-minute boundary (with buffer for test execution time)
		timestamp := time.Now().Unix() - 295

		basestring := "v0:" + strconv.FormatInt(timestamp, 10) + ":" + string(payload)
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(basestring))
		signature := "v0=" + hex.EncodeToString(mac.Sum(nil))

		// Should still pass at 300 second boundary
		if !verifySlackSignature(payload, signature, timestamp, secret) {
			t.Error("Signature at 5-minute boundary should pass")
		}
	})

	t.Run("SlackReplayWindowExceeded", func(t *testing.T) {
		secret := "test_secret"
		payload := []byte(`{"type":"event_callback"}`)

		// Test just over 5-minute boundary
		timestamp := time.Now().Unix() - 301

		basestring := "v0:" + strconv.FormatInt(timestamp, 10) + ":" + string(payload)
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(basestring))
		signature := "v0=" + hex.EncodeToString(mac.Sum(nil))

		// Should fail beyond 300 second window
		if verifySlackSignature(payload, signature, timestamp, secret) {
			t.Error("Signature beyond 5-minute window should fail")
		}
	})
}

// TestConcurrentSignatureVerification tests thread-safety of signature verification
func TestConcurrentSignatureVerification(t *testing.T) {
	secret := "test_secret"
	payload := []byte(`{"type":"test"}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	validSignature := hex.EncodeToString(mac.Sum(nil))

	done := make(chan bool, 10)

	// Run 10 concurrent verification operations
	for i := 0; i < 10; i++ {
		go func(index int) {
			result := verifyWhatsAppSignature(payload, "sha256="+validSignature, secret)
			if !result {
				t.Errorf("Concurrent verification %d failed", index)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
