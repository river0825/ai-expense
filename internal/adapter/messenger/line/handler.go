package line

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
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
	ctx := r.Context()

	// Verify signature
	signature := r.Header.Get("X-Line-Signature")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		lineLogger.ErrorContext(ctx, "failed to read LINE webhook body", "error", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	lineLogger.DebugContext(
		ctx,
		"received LINE webhook",
		"signature", maskToken(signature),
		"body_preview", previewText(string(body), 300),
	)

	if !h.verifySignature(signature, body) {
		lineLogger.WarnContext(ctx, "invalid LINE webhook signature", "signature", maskToken(signature))
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Parse events
	var event LineEvent
	if err := json.Unmarshal(body, &event); err != nil {
		lineLogger.ErrorContext(ctx, "failed to parse LINE webhook event", "error", err)
		http.Error(w, "Failed to parse event", http.StatusBadRequest)
		return
	}
	lineLogger.DebugContext(ctx, "parsed LINE webhook events", "count", len(event.Events))

	execCtx := context.Background()

	// Process each event
	for _, e := range event.Events {
		lineLogger.DebugContext(
			ctx,
			"processing LINE event",
			"type", e.Type,
			"message_type", e.Message.Type,
			"user_id", e.Source.UserID,
			"reply_token", maskToken(e.ReplyToken),
			"message_id", e.Message.ID,
		)
		if e.Type != "message" || e.Message.Type != "text" {
			lineLogger.DebugContext(
				ctx,
				"skipping unsupported LINE event",
				"type", e.Type,
				"message_type", e.Message.Type,
			)
			continue
		}

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
		resp, err := h.useCase.Execute(execCtx, userMsg)
		if err != nil {
			lineLogger.ErrorContext(
				ctx,
				"failed to handle LINE message",
				"error", err,
				"user_id", e.Source.UserID,
				"message_id", e.Message.ID,
				"text_preview", previewText(e.Message.Text, 120),
			)
			continue
		}
		if resp == nil {
			lineLogger.WarnContext(
				ctx,
				"usecase returned nil response",
				"user_id", e.Source.UserID,
				"message_id", e.Message.ID,
			)
			continue
		}
		lineLogger.DebugContext(
			ctx,
			"LINE response generated",
			"response_type", resp.Type,
			"text_len", len(resp.Text),
			"data_type", fmt.Sprintf("%T", resp.Data),
		)

		// Send reply as Flex Message
		if h.client != nil {
			locale := h.getUserLocale(execCtx, e.Source.UserID)
			h.sendFlexReply(execCtx, e.ReplyToken, resp, locale)
		} else {
			lineLogger.WarnContext(ctx, "LINE client is nil; response not sent")
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
	lineLogger.DebugContext(
		ctx,
		"building LINE reply",
		"response_type", resp.Type,
		"locale", locale,
		"reply_token", maskToken(replyToken),
	)

	switch resp.Type {
	case domain.ResponseTypeExpense:
		if data, ok := resp.Data.([]map[string]interface{}); ok {
			expenseData := convertToExpenseData(data)
			totalAmount, totalCurrency := extractTotal(data)
			flexBubble = flex.BuildExpenseBubble(expenseData, totalAmount, totalCurrency, locale)
			lineLogger.DebugContext(
				ctx,
				"built expense flex bubble",
				"items", len(expenseData),
				"total_amount", totalAmount,
				"total_currency", totalCurrency,
			)
		} else {
			lineLogger.WarnContext(ctx, "expense response data type mismatch", "data_type", fmt.Sprintf("%T", resp.Data))
		}
	case domain.ResponseTypeReport:
		if data, ok := resp.Data.(map[string]string); ok {
			flexBubble = flex.BuildReportBubble(data["link"], locale)
			lineLogger.DebugContext(ctx, "built report flex bubble", "link", data["link"])
		} else {
			lineLogger.WarnContext(ctx, "report response data type mismatch", "data_type", fmt.Sprintf("%T", resp.Data))
		}
	case domain.ResponseTypeError:
		flexBubble = flex.BuildErrorBubble(resp.Text, "", locale)
		lineLogger.DebugContext(ctx, "built error flex bubble")
	case domain.ResponseTypeInfo:
		flexBubble = flex.BuildInfoBubble(
			i18n.T(locale, "expense.none"),
			i18n.T(locale, "expense.none.hint"),
			locale,
		)
		lineLogger.DebugContext(ctx, "built info flex bubble")
	default:
		lineLogger.InfoContext(ctx, "unknown response type; will fallback to text if available", "response_type", resp.Type)
	}

	if flexBubble != nil {
		lineLogger.DebugContext(ctx, "sending flex reply", "alt_text_len", len(resp.Text), "reply_token", maskToken(replyToken))
		if err := h.client.SendFlexReply(ctx, replyToken, resp.Text, flexBubble); err != nil {
			lineLogger.ErrorContext(ctx, "failed to send flex reply; fallback to text", "error", err)
			if err := h.client.SendReply(ctx, replyToken, resp.Text); err != nil {
				lineLogger.ErrorContext(ctx, "fallback text reply failed", "error", err)
			}
		} else {
			lineLogger.InfoContext(ctx, "flex reply sent successfully", "reply_token", maskToken(replyToken))
		}
	} else if resp.Text != "" {
		lineLogger.DebugContext(ctx, "sending text reply because no flex bubble", "text_len", len(resp.Text))
		if err := h.client.SendReply(ctx, replyToken, resp.Text); err != nil {
			lineLogger.ErrorContext(ctx, "failed to send text reply", "error", err)
		}
	} else {
		lineLogger.WarnContext(ctx, "no flex bubble and empty response text; nothing sent")
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
