package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/riverlin/aiexpense/internal/domain"
)

var _ Service = (*GeminiAI)(nil)

const defaultGeminiModel = "gemini-2.5-flash-lite"

// GeminiAI implements the AI Service using Google Gemini API
type GeminiAI struct {
	apiKey string
	model  string
	// client *genai.Client // TODO: Initialize when Gemini SDK is available
}

// NewGeminiAI creates a new Gemini AI service
func NewGeminiAI(apiKey string, model string, costRepo domain.AICostRepository) (*GeminiAI, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}
	if model == "" {
		model = defaultGeminiModel
	}

	// TODO: Initialize Gemini client
	// client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
	// if err != nil {
	//     return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	// }

	return &GeminiAI{
		apiKey: apiKey,
		model:  model,
		// client: client,
	}, nil
}

// ParseExpense extracts expenses from natural language text
func (g *GeminiAI) ParseExpense(ctx context.Context, text string, userID string) (*ParseExpenseResponse, error) {
	log.Printf("DEBUG: GeminiAI.ParseExpense called with: %s", text)

	// Try Gemini API first
	resp, err := g.callGeminiAPI(ctx, text)
	if err == nil {
		// Note: Cost logging has moved to UseCase layer
		return resp, nil
	}

	log.Printf("WARN: Gemini API failed (using regex fallback): %v", err)

	// Fallback to regex - return zero token metadata since no API call was made
	expenses, err := g.parseExpenseRegex(text)
	if err != nil {
		return nil, err
	}

	return &ParseExpenseResponse{
		Expenses: expenses,
		Tokens: &TokenMetadata{
			InputTokens:  0,
			OutputTokens: 0,
			TotalTokens:  0,
		},
	}, nil
}

type geminiRequest struct {
	Contents         []geminiContent         `json:"contents"`
	GenerationConfig *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	ResponseMimeType string `json:"responseMimeType,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
	} `json:"usageMetadata"`
}

func (g *GeminiAI) sendGeminiRequest(ctx context.Context, prompt string) (*geminiResponse, string, error) {
	model := g.model
	if model == "" {
		model = defaultGeminiModel
	}
	url := "https://generativelanguage.googleapis.com/v1beta/models/" + model + ":generateContent?key=" + g.apiKey

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: &geminiGenerationConfig{
			ResponseMimeType: "application/json",
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}
	rawResponse := string(bodyBytes)

	if resp.StatusCode != http.StatusOK {
		return nil, rawResponse, fmt.Errorf("API error %d: %s", resp.StatusCode, rawResponse)
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(bodyBytes, &geminiResp); err != nil {
		return nil, rawResponse, fmt.Errorf("failed to decode response: %w", err)
	}

	return &geminiResp, rawResponse, nil
}

func (g *GeminiAI) callGeminiAPI(ctx context.Context, text string) (*ParseExpenseResponse, error) {
	prompt := fmt.Sprintf(`
You are an expense tracking assistant. Extract expenses from the following text.
Today is %s.

Return a JSON array of objects with these fields:
- description: string (what was bought)
- amount: number (price)
- suggested_category: string (Food, Transport, Shopping, Entertainment, Other)
- date: string (ISO 8601 format YYYY-MM-DD, resolve relative dates like "yesterday" based on today's date)

If the currency is not specified, assume TWD.
If no expenses are found, return an empty array [].

Text: %s
`, time.Now().Format("2006-01-02"), text)

	geminiResp, rawResp, err := g.sendGeminiRequest(ctx, prompt)
	if err != nil {
		return nil, err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	responseText := geminiResp.Candidates[0].Content.Parts[0].Text

	// Parse the JSON array from the response text
	var parsedItems []struct {
		Description       string  `json:"description"`
		Amount            float64 `json:"amount"`
		SuggestedCategory string  `json:"suggested_category"`
		Date              string  `json:"date"`
	}

	if err := json.Unmarshal([]byte(responseText), &parsedItems); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result JSON: %w", err)
	}

	var expenses []*domain.ParsedExpense
	for _, item := range parsedItems {
		var expenseDate time.Time
		if item.Date != "" {
			if parsedDate, err := time.Parse("2006-01-02", item.Date); err == nil {
				expenseDate = parsedDate
			} else {
				expenseDate = time.Now()
			}
		} else {
			expenseDate = time.Now()
		}

		expenses = append(expenses, &domain.ParsedExpense{
			Description:       item.Description,
			Amount:            item.Amount,
			SuggestedCategory: item.SuggestedCategory,
			Date:              expenseDate,
		})
	}

	// Extract token metadata from Gemini API response
	tokens := &TokenMetadata{
		InputTokens:  geminiResp.UsageMetadata.PromptTokenCount,
		OutputTokens: geminiResp.UsageMetadata.CandidatesTokenCount,
		TotalTokens:  geminiResp.UsageMetadata.PromptTokenCount + geminiResp.UsageMetadata.CandidatesTokenCount,
	}

	return &ParseExpenseResponse{
		Expenses:     expenses,
		Tokens:       tokens,
		SystemPrompt: prompt,
		RawResponse:  rawResp,
	}, nil
}

func (g *GeminiAI) callGeminiCategoryAPI(ctx context.Context, description string) (*SuggestCategoryResponse, error) {
	prompt := fmt.Sprintf(`
You are an expense tracking assistant. Categorize the following expense description into one of these categories:
- Food
- Transport
- Shopping
- Entertainment
- Other
- Health
- Education
- Bills

Description: %s

Return JUST the category name. Do not add any punctuation or explanation.
`, description)

	geminiResp, rawResp, err := g.sendGeminiRequest(ctx, prompt)
	if err != nil {
		return nil, err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	category := strings.TrimSpace(geminiResp.Candidates[0].Content.Parts[0].Text)

	// Clean up category string just in case
	category = strings.Trim(category, ".\"")

	tokens := &TokenMetadata{
		InputTokens:  geminiResp.UsageMetadata.PromptTokenCount,
		OutputTokens: geminiResp.UsageMetadata.CandidatesTokenCount,
		TotalTokens:  geminiResp.UsageMetadata.PromptTokenCount + geminiResp.UsageMetadata.CandidatesTokenCount,
	}

	return &SuggestCategoryResponse{
		Category:     category,
		Tokens:       tokens,
		SystemPrompt: prompt,
		RawResponse:  rawResp,
	}, nil
}

// parseExpenseRegex uses regex to extract expenses (fallback when AI unavailable)
func (g *GeminiAI) parseExpenseRegex(text string) ([]*domain.ParsedExpense, error) {
	var expenses []*domain.ParsedExpense

	// Helper to add expense
	addExpense := func(desc, amtStr string) {
		description := strings.TrimSpace(desc)
		if description == "" {
			return
		}
		// Clean description (remove trailing tokens if overlapping)
		description = strings.TrimSuffix(description, " ")

		amount, err := strconv.ParseFloat(amtStr, 64)
		if err != nil {
			return
		}

		expense := &domain.ParsedExpense{
			Description:       description,
			Amount:            amount,
			SuggestedCategory: "Other", // Default category
			// Date is left zero to let usecase handle relative date parsing
		}
		expenses = append(expenses, expense)
	}

	// Strategy: Try patterns from specific to general

	// Pattern 1: $ symbol (e.g., "lunch $10", "dinner$20")
	reDollar := regexp.MustCompile(`([^\d$]+?)\s*\$(\d+(?:\.\d{2})?)`)
	dollarMatches := reDollar.FindAllStringSubmatch(text, -1)

	// Pattern 2: '元' suffix (e.g., "早餐 10元", "午餐 100 元")
	reYuan := regexp.MustCompile(`(.*?)\s+(\d+(?:\.\d{2})?)\s*元`)
	yuanMatches := reYuan.FindAllStringSubmatch(text, -1)

	if len(dollarMatches) > 0 || len(yuanMatches) > 0 {
		for _, match := range dollarMatches {
			addExpense(match[1], match[2])
		}
		for _, match := range yuanMatches {
			addExpense(match[1], match[2])
		}
	} else {
		// Fallback: Loose space matching (e.g., "lunch 10")
		// Only use if no currency markers found to avoid duplicates or misparsing
		reSpace := regexp.MustCompile(`([^\d]+?)\s+(\d+(?:\.\d{2})?)(?:\s|$)`)
		matches := reSpace.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			addExpense(match[1], match[2])
		}
	}

	return expenses, nil
}

// SuggestCategory suggests a category based on description
func (g *GeminiAI) SuggestCategory(ctx context.Context, description string, userID string) (*SuggestCategoryResponse, error) {
	// Try Gemini API first
	resp, err := g.callGeminiCategoryAPI(ctx, description)
	if err == nil {
		return resp, nil
	}

	log.Printf("WARN: Gemini API failed for category suggestion (using fallback): %v", err)

	// Fallback to keyword matching (free, no API call)
	category := g.suggestCategoryKeywords(description)

	return &SuggestCategoryResponse{
		Category: category,
		Tokens: &TokenMetadata{
			InputTokens:  0,
			OutputTokens: 0,
			TotalTokens:  0,
		},
	}, nil
}

// suggestCategoryKeywords uses keyword matching for category suggestion (fallback)
func (g *GeminiAI) suggestCategoryKeywords(description string) string {
	description = strings.ToLower(description)

	foodKeywords := []string{"早餐", "午餐", "晚餐", "咖啡", "吃", "食物", "餐", "飯", "菜", "麵"}
	transportKeywords := []string{"加油", "公交", "計程車", "uber", "高鐵", "火車", "飛機", "停車", "油"}
	shoppingKeywords := []string{"買", "衣服", "鞋", "包", "購物", "店", "超市"}
	entertainmentKeywords := []string{"電影", "遊戲", "演唱會", "娛樂", "門票", "樂園"}

	for _, kw := range foodKeywords {
		if strings.Contains(description, kw) {
			return "Food"
		}
	}
	for _, kw := range transportKeywords {
		if strings.Contains(description, kw) {
			return "Transport"
		}
	}
	for _, kw := range shoppingKeywords {
		if strings.Contains(description, kw) {
			return "Shopping"
		}
	}
	for _, kw := range entertainmentKeywords {
		if strings.Contains(description, kw) {
			return "Entertainment"
		}
	}

	return "Other"
}
