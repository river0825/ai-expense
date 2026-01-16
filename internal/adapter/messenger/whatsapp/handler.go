package whatsapp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

// Handler handles WhatsApp webhook events
type Handler struct {
	appSecret string
	phone     string
	useCase   *WhatsAppUseCase
}

// NewHandler creates a new WhatsApp webhook handler
func NewHandler(appSecret, phoneNumber string, useCase *WhatsAppUseCase) *Handler {
	return &Handler{
		appSecret: appSecret,
		phone:     phoneNumber,
		useCase:   useCase,
	}
}

// WebhookPayload represents the webhook payload from WhatsApp
type WebhookPayload struct {
	Object string         `json:"object"`
	Entry  []WebhookEntry `json:"entry"`
}

// WebhookEntry represents an entry in the webhook payload
type WebhookEntry struct {
	ID        string          `json:"id"`
	Timestamp int64           `json:"timestamp"`
	Changes   []WebhookChange `json:"changes"`
}

// WebhookChange represents a change entry
type WebhookChange struct {
	Field string             `json:"field"`
	Value WebhookChangeValue `json:"value"`
}

// WebhookChangeValue represents the value of a change
type WebhookChangeValue struct {
	MessagingProduct string            `json:"messaging_product"`
	Metadata         MessageMetadata   `json:"metadata"`
	Messages         []IncomingMessage `json:"messages"`
	Statuses         []MessageStatus   `json:"statuses,omitempty"`
}

// MessageMetadata represents metadata from WhatsApp
type MessageMetadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

// IncomingMessage represents an incoming message from WhatsApp
type IncomingMessage struct {
	From        string             `json:"from"`
	ID          string             `json:"id"`
	Timestamp   string             `json:"timestamp"`
	Type        string             `json:"type"`
	Text        TextContent        `json:"text,omitempty"`
	Button      ButtonContent      `json:"button,omitempty"`
	Interactive InteractiveContent `json:"interactive,omitempty"`
}

// TextContent represents text message content
type TextContent struct {
	Body string `json:"body"`
}

// ButtonContent represents button message content
type ButtonContent struct {
	Text    string `json:"text"`
	Payload string `json:"payload"`
}

// InteractiveContent represents interactive message content
type InteractiveContent struct {
	Type        string      `json:"type"`
	ButtonReply ButtonReply `json:"button_reply,omitempty"`
}

// ButtonReply represents a button reply
type ButtonReply struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// MessageStatus represents a message status update
type MessageStatus struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Recipient string `json:"recipient_id,omitempty"`
}

// HandleWebhook handles incoming WhatsApp webhooks
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.handleVerification(w, r)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Verify webhook signature
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if !h.verifySignature(r.Header.Get("X-Hub-Signature-256"), string(body)) {
		log.Printf("Webhook signature verification failed")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("Error unmarshaling payload: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Process the payload
	h.processPayload(r, &payload)

	// Always return 200 to acknowledge receipt
	w.WriteHeader(http.StatusOK)
}

// handleVerification handles webhook verification from WhatsApp
func (h *Handler) handleVerification(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	challenge := params.Get("hub.challenge")
	token := params.Get("hub.verify_token")
	mode := params.Get("hub.mode")

	if mode != "subscribe" || token != "verify_token" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(challenge))
}

// verifySignature verifies the webhook signature
func (h *Handler) verifySignature(signature, payload string) bool {
	if signature == "" {
		return false
	}

	// Extract the hash from the signature header
	parts := strings.SplitN(signature, "=", 2)
	if len(parts) != 2 || parts[0] != "sha256" {
		return false
	}

	expectedHash := parts[1]

	// Calculate HMAC-SHA256
	hash := hmac.New(sha256.New, []byte(h.appSecret))
	hash.Write([]byte(payload))
	calculatedHash := hex.EncodeToString(hash.Sum(nil))

	return hmac.Equal([]byte(expectedHash), []byte(calculatedHash))
}

// processPayload processes the webhook payload
func (h *Handler) processPayload(r *http.Request, payload *WebhookPayload) {
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			if change.Field == "messages" {
				h.processMessages(r, &change.Value)
			}
		}
	}
}

// processMessages processes incoming messages
func (h *Handler) processMessages(r *http.Request, value *WebhookChangeValue) {
	for _, msg := range value.Messages {
		userID := msg.From
		var messageText string

		switch msg.Type {
		case "text":
			messageText = msg.Text.Body
		case "button":
			messageText = msg.Button.Payload
		case "interactive":
			if msg.Interactive.ButtonReply.Title != "" {
				messageText = msg.Interactive.ButtonReply.Title
			}
		default:
			log.Printf("Unsupported message type: %s", msg.Type)
			continue
		}

		if messageText == "" {
			log.Printf("Empty message from %s", userID)
			continue
		}

		// Handle the message asynchronously
		go func(uid, text string) {
			if err := h.useCase.HandleMessage(r.Context(), uid, text); err != nil {
				log.Printf("Error handling message from %s: %v", uid, err)
			}
		}(userID, messageText)
	}
}
