package usecase

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/riverlin/aiexpense/internal/ai"
	"github.com/riverlin/aiexpense/internal/domain"
)

// CreateExpenseUseCase handles creating new expenses with AI-powered category suggestion
type CreateExpenseUseCase struct {
	expenseRepo  domain.ExpenseRepository
	categoryRepo domain.CategoryRepository
	aiCostRepo   domain.AICostRepository
	pricingRepo  domain.PricingRepository
	aiService    ai.Service
	provider     string
	model        string
}

// NewCreateExpenseUseCase creates a new create expense use case
func NewCreateExpenseUseCase(
	expenseRepo domain.ExpenseRepository,
	categoryRepo domain.CategoryRepository,
	aiCostRepo domain.AICostRepository,
	pricingRepo domain.PricingRepository,
	aiService ai.Service,
) *CreateExpenseUseCase {
	return NewCreateExpenseUseCaseWithAIConfig(
		expenseRepo,
		categoryRepo,
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
		expenseRepo:  expenseRepo,
		categoryRepo: categoryRepo,
		aiCostRepo:   aiCostRepo,
		pricingRepo:  pricingRepo,
		aiService:    aiService,
		provider:     provider,
		model:        model,
	}
}

// CreateRequest represents a request to create an expense
type CreateRequest struct {
	UserID      string
	Description string
	Amount      float64
	CategoryID  *string
	Date        time.Time
}

// CreateResponse represents the response after creating an expense
type CreateResponse struct {
	ID       string
	Message  string
	Category string
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

	// Create expense
	expense := &domain.Expense{
		ID:          uuid.New().String(),
		UserID:      req.UserID,
		Description: req.Description,
		Amount:      req.Amount,
		CategoryID:  categoryID,
		ExpenseDate: req.Date,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := u.expenseRepo.Create(ctx, expense); err != nil {
		return nil, err
	}

	// Prepare response message
	message := req.Description + " " + formatAmount(req.Amount) + "元，已儲存"
	if categoryName != "" {
		message = req.Description + " " + formatAmount(req.Amount) + "元 [" + categoryName + "]，已儲存"
	}

	return &CreateResponse{
		ID:       expense.ID,
		Message:  message,
		Category: categoryName,
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
