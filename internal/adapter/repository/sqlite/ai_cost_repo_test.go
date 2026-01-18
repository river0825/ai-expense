package sqlite

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

// TestSQLiteAICostRepository integration tests
func TestSQLiteAICostRepository(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	// Ensure we are in project root for migrations
	if _, err := os.Stat("migrations"); os.IsNotExist(err) {
		// Attempt to move up to project root (from internal/adapter/repository/sqlite)
		os.Chdir("../../../..")
	}

	db, err := OpenDB(tmpfile.Name())
	if err != nil {
		t.Skipf("Skipping integration test: could not open database: %v (run from project root)", err)
		return
	}
	defer db.Close()

	userRepo := NewUserRepository(db)
	repo := NewAICostRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		UserID:        "cost_test_user",
		MessengerType: "line",
		CreatedAt:     time.Now(),
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	t.Run("CreateAndGetCostLog", func(t *testing.T) {
		log := &domain.AICostLog{
			ID:           "log_001",
			UserID:       "cost_test_user",
			Operation:    "parse_expense",
			Provider:     "gemini",
			Model:        "gemini-2.5-lite",
			InputTokens:  100,
			OutputTokens: 50,
			TotalTokens:  150,
			Cost:         0.00015,
			Currency:     "USD",
			CreatedAt:    time.Now(),
		}

		err := repo.Create(ctx, log)
		if err != nil {
			t.Fatalf("Failed to create cost log: %v", err)
		}

		logs, err := repo.GetByUserID(ctx, "cost_test_user", 10)
		if err != nil {
			t.Fatalf("Failed to get cost logs: %v", err)
		}

		if len(logs) == 0 {
			t.Fatal("Expected to retrieve at least 1 log")
		}

		retrieved := logs[0]
		if retrieved.ID != "log_001" {
			t.Errorf("Expected log ID 'log_001', got '%s'", retrieved.ID)
		}
		if retrieved.Cost != 0.00015 {
			t.Errorf("Expected cost 0.00015, got %f", retrieved.Cost)
		}
	})
}
