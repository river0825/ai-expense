package line

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/riverlin/aiexpense/internal/adapter/messenger/line/flex"
	"github.com/riverlin/aiexpense/internal/domain"
	"github.com/riverlin/aiexpense/internal/i18n"
)

// MessageProcessor defines the interface for processing messages
type MessageProcessor interface {
	Execute(ctx context.Context, msg *domain.UserMessage) (*domain.MessageResponse, error)
}

// Handler handles LINE bot webhook events
type Handler struct {
	channelSecret string
	useCase       MessageProcessor
	client        *Client
	userRepo      domain.UserRepository
}

// NewHandler creates a new LINE webhook handler
func NewHandler(channelSecret string, useCase MessageProcessor, client *Client, userRepo domain.UserRepository) *Handler {
	return &Handler{
		channelSecret: channelSecret,
		useCase:       useCase,
		client:        client,
		userRepo:      userRepo,
	}
}

// LineEvent represents a LINE messaging event
type LineEvent struct {
	Events []struct {
		Type    string `json:"type"`
		Message struct {
			ID   string `json:"id"`
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"message"`
		Source struct {
			Type   string `json:"type"`
			UserID string `json:"userId"`
		} `json:"source"`
		ReplyToken string `json:"replyToken"`
		Timestamp  int64  `json:"timestamp"`
	} `json:"events"`
}

// HandleWebhook processes incoming LINE webhook events
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Verify signature
	signature := r.Header.Get("X-Line-Signature")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// Debug log: Print webhook details
	log.Printf("[LINE Webhook] Signature: %s", signature)
	log.Printf("[LINE Webhook] Body: %s", string(body))

	if !h.verifySignature(signature, body) {
		log.Printf("[LINE Webhook] Invalid signature")
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Parse events
	var event LineEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "Failed to parse event", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Process each event
	for _, e := range event.Events {
		if e.Type != "message" || e.Message.Type != "text" {
			continue
		}

		log.Printf("[LINE Webhook] Processing message event from user %s: %s", e.Source.UserID, e.Message.Text)

		// Map to UserMessage
		userMsg := &domain.UserMessage{
			UserID:    e.Source.UserID,
			MessageID: e.Message.ID,
			Content:   e.Message.Text,
			Source:    "line",
			// Use event timestamp if available, otherwise Now
			Timestamp: time.Unix(e.Timestamp/1000, 0),
			Metadata: map[string]interface{}{
				"reply_token": e.ReplyToken,
			},
		}

		// Execute logic
		resp, err := h.useCase.Execute(ctx, userMsg)
		if err != nil {
			log.Printf("[LINE Webhook] Error handling message: %v", err)
			continue
		}

		// Send reply as Flex Message
		if resp != nil && h.client != nil {
			locale := h.getUserLocale(ctx, e.Source.UserID)
			h.sendFlexReply(ctx, e.ReplyToken, resp, locale)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) getUserLocale(ctx context.Context, userID string) string {
	if h.userRepo != nil {
		if user, err := h.userRepo.GetByID(ctx, userID); err == nil && user.Locale != "" {
			return user.Locale
		}
	}
	return i18n.DefaultLocale()
}

func (h *Handler) sendFlexReply(ctx context.Context, replyToken string, resp *domain.MessageResponse, locale string) {
	var flexBubble map[string]interface{}

	switch resp.Type {
	case domain.ResponseTypeExpense:
		if data, ok := resp.Data.([]map[string]interface{}); ok {
			expenseData := convertToExpenseData(data)
			totalAmount, totalCurrency := extractTotal(data)
			flexBubble = flex.BuildExpenseBubble(expenseData, totalAmount, totalCurrency, locale)
		}
	case domain.ResponseTypeReport:
		if data, ok := resp.Data.(map[string]string); ok {
			flexBubble = flex.BuildReportBubble(data["link"], locale)
		}
	case domain.ResponseTypeError:
		flexBubble = flex.BuildErrorBubble(resp.Text, "", locale)
	case domain.ResponseTypeInfo:
		flexBubble = flex.BuildInfoBubble(
			i18n.T(locale, "expense.none"),
			i18n.T(locale, "expense.none.hint"),
			locale,
		)
	}

	if flexBubble != nil {
		if err := h.client.SendFlexReply(ctx, replyToken, resp.Text, flexBubble); err != nil {
			log.Printf("[LINE Webhook] Failed to send flex reply: %v, falling back to text", err)
			if err := h.client.SendReply(ctx, replyToken, resp.Text); err != nil {
				log.Printf("[LINE Webhook] Fallback text reply also failed: %v", err)
			}
		} else {
			log.Printf("[LINE Webhook] Flex reply sent successfully")
		}
	} else if resp.Text != "" {
		if err := h.client.SendReply(ctx, replyToken, resp.Text); err != nil {
			log.Printf("[LINE Webhook] Failed to send reply: %v", err)
		}
	}
}

func convertToExpenseData(data []map[string]interface{}) []flex.ExpenseData {
	result := make([]flex.ExpenseData, 0, len(data))
	for _, d := range data {
		exp := flex.ExpenseData{
			Description:  stringVal(d, "description"),
			HomeAmount:   floatVal(d, "home_amount"),
			HomeCurrency: stringVal(d, "home_currency"),
			Category:     stringVal(d, "category"),
			Account:      stringVal(d, "account"),
		}
		if orig := floatVal(d, "original_amount"); orig > 0 {
			exp.OriginalAmount = orig
			exp.OriginalCurrency = stringVal(d, "currency")
		}
		if dt, ok := d["date"].(time.Time); ok {
			exp.Date = dt
		}
		if exp.HomeCurrency == "" {
			exp.HomeCurrency = "TWD"
		}
		if exp.HomeAmount == 0 {
			exp.HomeAmount = exp.OriginalAmount
		}
		result = append(result, exp)
	}
	return result
}

func extractTotal(data []map[string]interface{}) (float64, string) {
	total := 0.0
	currency := "TWD"
	for _, d := range data {
		total += floatVal(d, "home_amount")
		if c := stringVal(d, "home_currency"); c != "" {
			currency = c
		}
	}
	return total, currency
}

func stringVal(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func floatVal(m map[string]interface{}, key string) float64 {
	switch v := m[key].(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	default:
		return 0
	}
}

// verifySignature verifies the LINE webhook signature
func (h *Handler) verifySignature(signature string, body []byte) bool {
	hash := hmac.New(sha256.New, []byte(h.channelSecret))
	hash.Write(body)
	computed := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	return strings.EqualFold(signature, computed)
}
