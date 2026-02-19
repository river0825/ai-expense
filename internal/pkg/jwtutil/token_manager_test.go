package jwtutil

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestTokenManager_GenerateUserToken(t *testing.T) {
	secret := "test-secret"
	manager := NewTokenManager(secret)
	userID := "user-123"
	duration := time.Hour

	token, err := manager.GenerateUserToken(userID, duration)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := manager.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims["sub"])

	extractedID, err := manager.GetUserIDFromClaims(claims)
	assert.NoError(t, err)
	assert.Equal(t, userID, extractedID)
}

func TestTokenManager_GenerateReportToken(t *testing.T) {
	secret := "test-secret"
	manager := NewTokenManager(secret)
	userID := "user-123"
	duration := time.Hour

	token, err := manager.GenerateReportToken(userID, duration)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := manager.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims["sub"])
	assert.Equal(t, "report_access", claims["type"])
}

func TestTokenManager_ValidateToken_InvalidToken(t *testing.T) {
	secret := "test-secret"
	manager := NewTokenManager(secret)

	_, err := manager.ValidateToken("invalid-token")
	assert.Error(t, err)
}

func TestTokenManager_ValidateToken_WrongSecret(t *testing.T) {
	secret1 := "secret-1"
	manager1 := NewTokenManager(secret1)
	token, _ := manager1.GenerateUserToken("user-1", time.Hour)

	secret2 := "secret-2"
	manager2 := NewTokenManager(secret2)
	_, err := manager2.ValidateToken(token)
	assert.Error(t, err)
}

func TestTokenManager_GetUserIDFromClaims_MissingSub(t *testing.T) {
	manager := NewTokenManager("secret")
	claims := jwt.MapClaims{
		"foo": "bar",
	}

	_, err := manager.GetUserIDFromClaims(claims)
	assert.Error(t, err)
}
