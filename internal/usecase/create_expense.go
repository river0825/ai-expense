package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/riverlin/aiexpense/internal/ai"
	"github.com/riverlin/aiexpense/internal/domain"
)

// CreateExpenseUseCase handles creating new expenses with AI-powered category suggestion
type CreateExpenseUseCase struct {
	expenseRepo     domain.ExpenseRepository
	categoryRepo    domain.CategoryRepository
	userRepo        domain.UserRepository
	exchangeRateSvc domain.ExchangeRateService
	aiCostRepo      domain.AICostRepository
	pricingRepo     domain.PricingRepository
	aiService       ai.Service
	provider        string
	model           string
}

// NewCreateExpenseUseCase creates a new create expense use case
func NewCreateExpenseUseCase(
	expenseRepo domain.ExpenseRepository,
	categoryRepo domain.CategoryRepository,
	userRepo domain.UserRepository,
	exchangeRateSvc domain.ExchangeRateService,
	aiCostRepo domain.AICostRepository,
	pricingRepo domain.PricingRepository,
	aiService ai.Service,
) *CreateExpenseUseCase {
	return NewCreateExpenseUseCaseWithAIConfig(
		expenseRepo,
		categoryRepo,
		userRepo,
		exchangeRateSvc,
		aiCostRepo,
		pricingRepo,
		aiService,
		"gemini",
		"gemini-2.5-flash-lite",
	)
}

// NewCreateExpenseUseCaseWithAIConfig creates a new create expense use case with provider/model for cost logging
func NewCreateExpenseUseCaseWithAIConfig(
	expenseRepo domain.ExpenseRepository,
	categoryRepo domain.CategoryRepository,
	userRepo domain.UserRepository,
	exchangeRateSvc domain.ExchangeRateService,
	aiCostRepo domain.AICostRepository,
	pricingRepo domain.PricingRepository,
	aiService ai.Service,
	provider string,
	model string,
) *CreateExpenseUseCase {
	if provider == "" {
		provider = "gemini"
	}
	if model == "" {
		model = "gemini-2.5-flash-lite"
	}
	return &CreateExpenseUseCase{
		expenseRepo:     expenseRepo,
		categoryRepo:    categoryRepo,
		userRepo:        userRepo,
		exchangeRateSvc: exchangeRateSvc,
		aiCostRepo:      aiCostRepo,
		pricingRepo:     pricingRepo,
		aiService:       aiService,
		provider:        provider,
		model:           model,
	}
}

// CreateRequest represents a request to create an expense
type CreateRequest struct {
	UserID           string
	Description      string
	Amount           float64
	Currency         string
	CurrencyOriginal string
	ConvertedAmount  float64
	HomeCurrency     string
	ExchangeRate     float64
	CategoryID       *string
	Account          string
	Date             time.Time
}

// CreateResponse represents the response after creating an expense
type CreateResponse struct {
	ID             string
	Message        string
	Category       string
	OriginalAmount float64
	Currency       string
	HomeAmount     float64
	HomeCurrency   string
	ExchangeRate   float64
	Account        string
}

// Execute creates a new expense
func (u *CreateExpenseUseCase) Execute(ctx context.Context, req *CreateRequest) (*CreateResponse, error) {
	// If no category is specified, get AI suggestion
	var categoryID *string
	var categoryName string

	if req.CategoryID != nil {
		categoryID = req.CategoryID
		// Get category name for response
		category, _ := u.categoryRepo.GetByID(ctx, *req.CategoryID)
		if category != nil {
			categoryName = category.Name
			log.Printf("Expense created with manual category: %s (ID: %s)", categoryName, *req.CategoryID)
		}
	} else {
		// Get AI suggestion
		resp, err := u.aiService.SuggestCategory(ctx, req.Description, req.UserID)
		if err == nil && resp != nil {
			log.Printf("AI suggested category: %s for description: %s", resp.Category, req.Description)

			// Log AI cost
			if resp.Tokens != nil {
				go func() {
					// Create background context for logging to not block response
					logCtx := context.Background()
					cost := 0.0
					provider := u.provider
					model := u.model

					// Calculate cost if pricing is available
					if u.pricingRepo != nil {
						pricing, err := u.pricingRepo.GetByProviderAndModel(logCtx, provider, model)
						if err == nil && pricing != nil {
							cost = pricing.GetCost(resp.Tokens.InputTokens, resp.Tokens.OutputTokens)
						}
					}

					costLog := &domain.AICostLog{
						ID:           uuid.New().String(),
						UserID:       req.UserID,
						Operation:    "suggest_category",
						Provider:     provider,
						Model:        model,
						InputTokens:  resp.Tokens.InputTokens,
						OutputTokens: resp.Tokens.OutputTokens,
						TotalTokens:  resp.Tokens.TotalTokens,
						Cost:         cost,
						Currency:     "USD",
						CreatedAt:    time.Now(),
					}

					if u.aiCostRepo != nil {
						if err := u.aiCostRepo.Create(logCtx, costLog); err != nil {
							log.Printf("Failed to log AI cost: %v", err)
						}
					}
				}()
			}

			// Find category by name
			categories, _ := u.categoryRepo.GetByUserID(ctx, req.UserID)
			for _, cat := range categories {
				if cat.Name == resp.Category {
					categoryID = &cat.ID
					categoryName = cat.Name
					break
				}
			}
		}
	}

	// Handle default account
	account := req.Account
	if account == "" {
		account = "Cash"
	}

	// Create expense
	originalAmount := req.Amount
	homeCurrency := u.resolveHomeCurrency(ctx, req.UserID, normalizeCurrency(req.HomeCurrency))
	currency := normalizeCurrency(req.Currency)
	if currency == "" {
		currency = homeCurrency
	}
	homeAmount := req.ConvertedAmount
	exchangeRate := req.ExchangeRate
	if homeAmount <= 0 {
		if u.exchangeRateSvc != nil && currency != homeCurrency {
			converted, rate, err := u.exchangeRateSvc.Convert(ctx, originalAmount, currency, homeCurrency, req.Date)
			if err == nil {
				homeAmount = converted
				exchangeRate = rate
			} else {
				log.Printf("WARN: failed currency conversion %s->%s: %v", currency, homeCurrency, err)
				homeAmount = originalAmount
				exchangeRate = 1.0
			}
		} else {
			homeAmount = originalAmount
			if exchangeRate == 0 {
				exchangeRate = 1.0
			}
		}
	}
	if exchangeRate == 0 {
		exchangeRate = 1.0
	}

	expense := &domain.Expense{
		ID:             uuid.New().String(),
		UserID:         req.UserID,
		Description:    req.Description,
		OriginalAmount: originalAmount,
		Currency:       currency,
		HomeAmount:     homeAmount,
		HomeCurrency:   homeCurrency,
		ExchangeRate:   exchangeRate,
		CategoryID:     categoryID,
		Account:        account,
		ExpenseDate:    req.Date,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	expense.Amount = expense.HomeAmount

	if err := u.expenseRepo.Create(ctx, expense); err != nil {
		return nil, err
	}

	// Prepare response message
	message := buildCreateMessage(req.Description, originalAmount, currency, homeAmount, homeCurrency, categoryName)

	return &CreateResponse{
		ID:             expense.ID,
		Message:        message,
		Category:       categoryName,
		OriginalAmount: originalAmount,
		Currency:       currency,
		HomeAmount:     homeAmount,
		HomeCurrency:   homeCurrency,
		ExchangeRate:   exchangeRate,
		Account:        account,
	}, nil
}

// formatAmount formats amount for display
func formatAmount(amount float64) string {
	if amount == float64(int64(amount)) {
		return formatInt(int64(amount))
	}
	return formatFloat(amount)
}

// formatInt formats integer
func formatInt(n int64) string {
	if n == 0 {
		return "0"
	}
	result := ""
	if n < 0 {
		result = "-"
		n = -n
	}
	for n > 0 {
		result = string('0'+byte(n%10)) + result
		n /= 10
	}
	return result
}

// formatFloat formats float
func formatFloat(f float64) string {
	s := ""
	if f < 0 {
		s = "-"
		f = -f
	}

	// Format integer part
	intPart := int64(f)
	s += formatInt(intPart)

	// Format decimal part
	decPart := f - float64(intPart)
	if decPart > 0 {
		s += "."
		for i := 0; i < 2 && decPart > 0; i++ {
			decPart *= 10
			digit := int64(decPart)
			s += string('0' + byte(digit))
			decPart -= float64(digit)
		}
	}

	return s
}

func normalizeCurrency(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func (u *CreateExpenseUseCase) resolveHomeCurrency(ctx context.Context, userID, fallback string) string {
	if fallback != "" {
		return fallback
	}
	if u.userRepo == nil {
		return "TWD"
	}
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil || user.HomeCurrency == "" {
		return "TWD"
	}
	return strings.ToUpper(user.HomeCurrency)
}

func buildCreateMessage(description string, originalAmount float64, currency string, homeAmount float64, homeCurrency string, categoryName string) string {
	var message string
	if currency != "" && currency != homeCurrency {
		message = fmt.Sprintf("%s %s %s (≈ %s %s)", description, formatAmount(originalAmount), currency, formatAmount(homeAmount), homeCurrency)
	} else {
		message = fmt.Sprintf("%s %s %s", description, formatAmount(homeAmount), homeCurrency)
	}
	if categoryName != "" {
		message = fmt.Sprintf("%s [%s]", message, categoryName)
	}
	return message + "，已儲存"
}
