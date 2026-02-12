# LINE Flex Messages + i18n Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace plain text LINE bot responses with styled Flex Messages and add backend i18n (zh-TW default, en) using go:embed translation files.

**Architecture:** Three independent layers built bottom-up: i18n package (no dependencies) -> domain MessageResponse.Type field -> LINE Flex builders + client + handler integration. LINE profile language detection on auto-signup.

**Tech Stack:** Go 1.24, go:embed, LINE Messaging API (Flex Messages), LINE Profile API

**Design doc:** `docs/plans/2026-02-10-line-flex-messages-i18n-design.md`

---

### Task 1: i18n Package — Translation Files

**Files:**
- Create: `internal/i18n/locales/zh-TW.json`
- Create: `internal/i18n/locales/en.json`

**Step 1: Create zh-TW translation file**

Create `internal/i18n/locales/zh-TW.json`:

```json
{
  "expense.recorded": "已記錄 {count} 筆支出，合計：{amount} {currency}",
  "expense.item": "{description}: {amount} {currency}",
  "expense.item.detail": "{category} · {date} · {account}",
  "expense.item.converted": "≈ {originalAmount} {originalCurrency}",
  "expense.none": "訊息中未偵測到任何支出",
  "expense.none.hint": "請嘗試輸入如：「午餐 便當 85元」",
  "expense.count": "{count} 筆支出已成功記錄",
  "report.title": "支出報告",
  "report.description": "您的個人支出報告已準備好",
  "report.validity": "連結有效期 5 分鐘",
  "report.button": "查看報告",
  "report.error": "抱歉，無法產生報告連結，請稍後再試。",
  "error.title": "處理失敗",
  "error.signup": "使用者註冊失敗：{error}",
  "error.parse": "訊息解析失敗：{error}",
  "flex.app_name": "AIExpense",
  "flex.total": "合計"
}
```

**Step 2: Create en translation file**

Create `internal/i18n/locales/en.json`:

```json
{
  "expense.recorded": "Recorded {count} expense(s), total: {amount} {currency}",
  "expense.item": "{description}: {amount} {currency}",
  "expense.item.detail": "{category} · {date} · {account}",
  "expense.item.converted": "≈ {originalAmount} {originalCurrency}",
  "expense.none": "No expenses detected in message",
  "expense.none.hint": "Try something like: \"Lunch bento $85\"",
  "expense.count": "{count} expense(s) recorded successfully",
  "report.title": "Expense Report",
  "report.description": "Your personal expense report is ready",
  "report.validity": "Link valid for 5 minutes",
  "report.button": "View Report",
  "report.error": "Sorry, couldn't generate the report link. Please try again later.",
  "error.title": "Error",
  "error.signup": "Failed to sign up user: {error}",
  "error.parse": "Failed to parse message: {error}",
  "flex.app_name": "AIExpense",
  "flex.total": "Total"
}
```

**Step 3: Commit**

```bash
git add internal/i18n/locales/zh-TW.json internal/i18n/locales/en.json
git commit -m "feat(i18n): add zh-TW and en translation files"
```

---

### Task 2: i18n Package — Translation Engine

**Files:**
- Create: `internal/i18n/i18n.go`
- Create: `internal/i18n/i18n_test.go`

**Step 1: Write failing tests**

Create `internal/i18n/i18n_test.go`:

```go
package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestT_ReturnsZhTWByDefault(t *testing.T) {
	result := T("zh-TW", "expense.none")
	assert.Equal(t, "訊息中未偵測到任何支出", result)
}

func TestT_ReturnsEnglish(t *testing.T) {
	result := T("en", "expense.none")
	assert.Equal(t, "No expenses detected in message", result)
}

func TestT_FallsBackToZhTW(t *testing.T) {
	// Unknown locale falls back to zh-TW
	result := T("ja", "expense.none")
	assert.Equal(t, "訊息中未偵測到任何支出", result)
}

func TestT_ReturnsKeyIfNotFound(t *testing.T) {
	result := T("zh-TW", "nonexistent.key")
	assert.Equal(t, "nonexistent.key", result)
}

func TestTf_SubstitutesParams(t *testing.T) {
	result := Tf("zh-TW", "expense.recorded", map[string]string{
		"count":    "3",
		"amount":   "1,250",
		"currency": "TWD",
	})
	assert.Equal(t, "已記錄 3 筆支出，合計：1,250 TWD", result)
}

func TestTf_EnglishWithParams(t *testing.T) {
	result := Tf("en", "expense.recorded", map[string]string{
		"count":    "2",
		"amount":   "500",
		"currency": "USD",
	})
	assert.Equal(t, "Recorded 2 expense(s), total: 500 USD", result)
}

func TestDefaultLocale(t *testing.T) {
	assert.Equal(t, "zh-TW", DefaultLocale())
}

func TestSupportedLocales(t *testing.T) {
	locales := SupportedLocales()
	assert.Contains(t, locales, "zh-TW")
	assert.Contains(t, locales, "en")
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/i18n/ -v`
Expected: FAIL (package doesn't exist yet)

**Step 3: Implement the i18n engine**

Create `internal/i18n/i18n.go`:

```go
package i18n

import (
	"embed"
	"encoding/json"
	"strings"
	"sync"
)

//go:embed locales/*.json
var localeFS embed.FS

const defaultLocale = "zh-TW"

var (
	translations map[string]map[string]string
	once         sync.Once
)

func load() {
	once.Do(func() {
		translations = make(map[string]map[string]string)
		entries, err := localeFS.ReadDir("locales")
		if err != nil {
			return
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !strings.HasSuffix(name, ".json") {
				continue
			}
			locale := strings.TrimSuffix(name, ".json")
			data, err := localeFS.ReadFile("locales/" + name)
			if err != nil {
				continue
			}
			var m map[string]string
			if err := json.Unmarshal(data, &m); err != nil {
				continue
			}
			translations[locale] = m
		}
	})
}

// T returns a translated string for the given locale and key.
// Falls back to zh-TW if key not found in requested locale.
func T(locale, key string) string {
	load()
	if m, ok := translations[locale]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	// Fallback to default locale
	if locale != defaultLocale {
		if m, ok := translations[defaultLocale]; ok {
			if v, ok := m[key]; ok {
				return v
			}
		}
	}
	return key
}

// Tf returns a translated string with named parameter substitution.
// Params like {count}, {amount} are replaced from the map.
func Tf(locale, key string, params map[string]string) string {
	tmpl := T(locale, key)
	if len(params) == 0 {
		return tmpl
	}
	oldNew := make([]string, 0, len(params)*2)
	for k, v := range params {
		oldNew = append(oldNew, "{"+k+"}", v)
	}
	return strings.NewReplacer(oldNew...).Replace(tmpl)
}

// SupportedLocales returns all available locale codes.
func SupportedLocales() []string {
	load()
	locales := make([]string, 0, len(translations))
	for k := range translations {
		locales = append(locales, k)
	}
	return locales
}

// DefaultLocale returns "zh-TW".
func DefaultLocale() string {
	return defaultLocale
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/i18n/ -v`
Expected: PASS (all 8 tests)

**Step 5: Commit**

```bash
git add internal/i18n/i18n.go internal/i18n/i18n_test.go
git commit -m "feat(i18n): add translation engine with go:embed"
```

---

### Task 3: Domain — Add Type to MessageResponse

**Files:**
- Modify: `internal/domain/messenger.go`
- Modify: `internal/usecase/process_message.go`
- Modify: `internal/usecase/process_message_test.go`

**Step 1: Add response type constants and Type field to MessageResponse**

In `internal/domain/messenger.go`, replace the `MessageResponse` struct:

```go
// MessageResponseType represents the type of bot response
const (
	ResponseTypeExpense = "expense"
	ResponseTypeReport  = "report"
	ResponseTypeError   = "error"
	ResponseTypeInfo    = "info"
)

// MessageResponse represents a standard response to be sent back to the user
type MessageResponse struct {
	Type string      `json:"type"`           // "expense", "report", "error", "info"
	Text string      `json:"text"`           // Plain text fallback
	Data interface{} `json:"data,omitempty"` // Structured data for rich rendering
}
```

**Step 2: Update ProcessMessageUseCase to set Type on all responses**

In `internal/usecase/process_message.go`, update each `return &domain.MessageResponse{...}`:

1. **Signup error** (line ~98): Add `Type: domain.ResponseTypeError`
2. **Report link success** (line ~115): Add `Type: domain.ResponseTypeReport`, add `Data: map[string]string{"link": link}`
3. **Report link error** (line ~115): Add `Type: domain.ResponseTypeError`
4. **Parse error** (line ~125): Add `Type: domain.ResponseTypeError`
5. **No expenses** (line ~136): Add `Type: domain.ResponseTypeInfo`
6. **Expenses recorded** (line ~220): Add `Type: domain.ResponseTypeExpense` (Data already set)

**Step 3: Update process_message_test.go assertions**

In the existing tests, add `Type` assertions:

- "Success - Single Expense" test: add `assert.Equal(t, domain.ResponseTypeExpense, resp.Type)`
- "Failure - Parse Error" test: add `assert.Equal(t, domain.ResponseTypeError, resp.Type)`

**Step 4: Run tests**

Run: `go test ./internal/usecase/ -run TestProcessMessage -v`
Expected: PASS

Run: `go test ./internal/adapter/messenger/line/ -v`
Expected: PASS (existing tests should still work — Type is a new field, not breaking)

**Step 5: Commit**

```bash
git add internal/domain/messenger.go internal/usecase/process_message.go internal/usecase/process_message_test.go
git commit -m "feat(domain): add Type field to MessageResponse"
```

---

### Task 4: LINE Client — SendFlexReply Method

**Files:**
- Modify: `internal/adapter/messenger/line/client.go`

**Step 1: Add Flex Message types and SendFlexReply**

Add to `internal/adapter/messenger/line/client.go`:

```go
// FlexReplyRequest represents the request to send a Flex Message reply
type FlexReplyRequest struct {
	ReplyToken string        `json:"replyToken"`
	Messages   []FlexMessage `json:"messages"`
}

// FlexMessage represents a LINE Flex Message
type FlexMessage struct {
	Type     string      `json:"type"`
	AltText  string      `json:"altText"`
	Contents interface{} `json:"contents"`
}

// SendFlexReply sends a Flex Message reply to a user via LINE Messaging API
func (c *Client) SendFlexReply(ctx context.Context, replyToken, altText string, contents interface{}) error {
	req := FlexReplyRequest{
		ReplyToken: replyToken,
		Messages: []FlexMessage{
			{
				Type:     "flex",
				AltText:  altText,
				Contents: contents,
			},
		},
	}

	payload, err := json.Marshal(req)
	if err != nil {
		log.Printf("Error marshaling flex request: %v", err)
		return fmt.Errorf("failed to marshal flex request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/reply", c.apiURL), bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.channelToken))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("Error sending flex message to LINE: %v", err)
		return fmt.Errorf("failed to send flex message: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		log.Printf("[LINE API Error] Status: %d, Body: %s", resp.StatusCode, string(body))
		var apiResp LineAPIResponse
		if err := json.Unmarshal(body, &apiResp); err == nil && apiResp.Message != "" {
			return fmt.Errorf("line api error: %s (status: %d)", apiResp.Message, resp.StatusCode)
		}
		return fmt.Errorf("line api error: status %d, body: %s", resp.StatusCode, string(body))
	}

	log.Printf("[LINE] Flex message sent to reply token %s", replyToken)
	return nil
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/adapter/messenger/line/...`
Expected: BUILD SUCCESS

**Step 3: Commit**

```bash
git add internal/adapter/messenger/line/client.go
git commit -m "feat(line): add SendFlexReply method to LINE client"
```

---

### Task 5: LINE Profile API

**Files:**
- Create: `internal/adapter/messenger/line/profile.go`
- Create: `internal/adapter/messenger/line/profile_test.go`

**Step 1: Write failing test**

Create `internal/adapter/messenger/line/profile_test.go`:

```go
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
		httpClient:   server.Client(),
		profileURL:   server.URL + "/v2/bot/profile",
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
		httpClient:   server.Client(),
		profileURL:   server.URL + "/v2/bot/profile",
	}

	_, err := client.GetProfile(context.Background(), "U_INVALID")
	assert.Error(t, err)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/adapter/messenger/line/ -run TestClient_GetProfile -v`
Expected: FAIL (GetProfile and profileURL don't exist)

**Step 3: Update Client struct and NewClient**

In `internal/adapter/messenger/line/client.go`, add `profileURL` field to `Client`:

```go
type Client struct {
	channelToken string
	apiURL       string
	profileURL   string
	httpClient   *http.Client
}
```

Update `NewClient`:

```go
func NewClient(channelToken string) (*Client, error) {
	if channelToken == "" {
		return nil, fmt.Errorf("channel token is required")
	}
	return &Client{
		channelToken: channelToken,
		apiURL:       "https://api.line.me/v2/bot/message",
		profileURL:   "https://api.line.me/v2/bot/profile",
		httpClient:   &http.Client{},
	}, nil
}
```

**Step 4: Create profile.go**

Create `internal/adapter/messenger/line/profile.go`:

```go
package line

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// LineProfile represents a LINE user profile
type LineProfile struct {
	DisplayName   string `json:"displayName"`
	Language      string `json:"language"`
	PictureURL    string `json:"pictureUrl"`
	StatusMessage string `json:"statusMessage"`
}

// GetProfile fetches the user's LINE profile including language.
func (c *Client) GetProfile(ctx context.Context, userID string) (*LineProfile, error) {
	url := fmt.Sprintf("%s/%s", c.profileURL, userID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create profile request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.channelToken))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("Error fetching LINE profile: %v", err)
		return nil, fmt.Errorf("failed to fetch profile: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[LINE Profile API Error] Status: %d, Body: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("line profile api error: status %d", resp.StatusCode)
	}

	var profile LineProfile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	return &profile, nil
}
```

**Step 5: Run tests**

Run: `go test ./internal/adapter/messenger/line/ -run TestClient_GetProfile -v`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/adapter/messenger/line/client.go internal/adapter/messenger/line/profile.go internal/adapter/messenger/line/profile_test.go
git commit -m "feat(line): add LINE Profile API for language detection"
```

---

### Task 6: Flex Message Builders — Expense Template

**Files:**
- Create: `internal/adapter/messenger/line/flex/expense.go`
- Create: `internal/adapter/messenger/line/flex/expense_test.go`

**Step 1: Write failing test**

Create `internal/adapter/messenger/line/flex/expense_test.go`:

```go
package flex

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildExpenseBubble_SingleExpense(t *testing.T) {
	expenses := []ExpenseData{
		{
			Description:  "午餐 便當",
			HomeAmount:   85,
			HomeCurrency: "TWD",
			Category:     "餐飲",
			Account:      "現金",
			Date:         time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
		},
	}

	bubble := BuildExpenseBubble(expenses, 85, "TWD", "zh-TW")

	require.NotNil(t, bubble)
	// Should be a bubble type
	assert.Equal(t, "bubble", bubble["type"])
	// Should have header, body, footer
	assert.NotNil(t, bubble["header"])
	assert.NotNil(t, bubble["body"])
	assert.NotNil(t, bubble["footer"])
}

func TestBuildExpenseBubble_MultiCurrency(t *testing.T) {
	expenses := []ExpenseData{
		{
			Description:      "Coffee",
			HomeAmount:       150,
			HomeCurrency:     "TWD",
			OriginalAmount:   5,
			OriginalCurrency: "USD",
			Category:         "Food",
			Account:          "Credit Card",
			Date:             time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
		},
	}

	bubble := BuildExpenseBubble(expenses, 150, "TWD", "en")

	require.NotNil(t, bubble)
	assert.Equal(t, "bubble", bubble["type"])
}

func TestBuildExpenseBubble_English(t *testing.T) {
	expenses := []ExpenseData{
		{
			Description:  "Lunch",
			HomeAmount:   20,
			HomeCurrency: "USD",
			Category:     "Food",
			Account:      "Cash",
			Date:         time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
		},
	}

	bubble := BuildExpenseBubble(expenses, 20, "USD", "en")

	require.NotNil(t, bubble)
	assert.Equal(t, "bubble", bubble["type"])
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/adapter/messenger/line/flex/ -v`
Expected: FAIL (package doesn't exist)

**Step 3: Implement expense builder**

Create `internal/adapter/messenger/line/flex/expense.go`:

```go
package flex

import (
	"fmt"
	"time"

	"github.com/riverlin/aiexpense/internal/i18n"
)

// ExpenseData holds the data needed to render a single expense in a Flex Message
type ExpenseData struct {
	Description      string
	HomeAmount       float64
	HomeCurrency     string
	OriginalAmount   float64
	OriginalCurrency string
	Category         string
	Account          string
	Date             time.Time
}

// BuildExpenseBubble creates a LINE Flex Message bubble for expense confirmation.
func BuildExpenseBubble(expenses []ExpenseData, totalAmount float64, totalCurrency, locale string) map[string]interface{} {
	// Header
	header := map[string]interface{}{
		"type":            "box",
		"layout":          "vertical",
		"backgroundColor": "#059669",
		"paddingAll":      "16px",
		"contents": []interface{}{
			map[string]interface{}{
				"type":  "text",
				"text":  i18n.T(locale, "flex.app_name"),
				"color": "#FFFFFF",
				"size":  "xs",
			},
			map[string]interface{}{
				"type":   "text",
				"text":   i18n.Tf(locale, "expense.recorded", map[string]string{"count": fmt.Sprintf("%d", len(expenses)), "amount": formatAmount(totalAmount), "currency": totalCurrency}),
				"color":  "#FFFFFF",
				"size":   "sm",
				"margin": "sm",
				"wrap":   true,
			},
			map[string]interface{}{
				"type":   "text",
				"text":   fmt.Sprintf("%s %s %s", i18n.T(locale, "flex.total"), formatAmount(totalAmount), totalCurrency),
				"color":  "#FFFFFF",
				"size":   "xxl",
				"weight": "bold",
				"margin": "sm",
			},
		},
	}

	// Body: expense items
	bodyContents := []interface{}{}
	for idx, exp := range expenses {
		if idx > 0 {
			bodyContents = append(bodyContents, map[string]interface{}{
				"type":  "separator",
				"color": "#E2E8F0",
			})
		}

		// Description + amount row
		amountText := fmt.Sprintf("%s %s", formatAmount(exp.HomeAmount), exp.HomeCurrency)
		row := map[string]interface{}{
			"type":   "box",
			"layout": "horizontal",
			"margin": "md",
			"contents": []interface{}{
				map[string]interface{}{
					"type":   "text",
					"text":   exp.Description,
					"size":   "sm",
					"color":  "#1E293B",
					"weight": "bold",
					"flex":   4,
				},
				map[string]interface{}{
					"type":   "text",
					"text":   amountText,
					"size":   "sm",
					"color":  "#059669",
					"weight": "bold",
					"align":  "end",
					"flex":   3,
				},
			},
		}
		bodyContents = append(bodyContents, row)

		// Detail row: category · date · account
		detailParts := []string{}
		if exp.Category != "" {
			detailParts = append(detailParts, exp.Category)
		}
		detailParts = append(detailParts, exp.Date.Format("2006-01-02"))
		if exp.Account != "" {
			detailParts = append(detailParts, exp.Account)
		}
		detailText := ""
		for i, p := range detailParts {
			if i > 0 {
				detailText += " · "
			}
			detailText += p
		}
		detail := map[string]interface{}{
			"type":   "text",
			"text":   detailText,
			"size":   "xxs",
			"color":  "#64748B",
			"margin": "xs",
		}
		bodyContents = append(bodyContents, detail)

		// Multi-currency indicator
		if exp.OriginalCurrency != "" && exp.OriginalCurrency != exp.HomeCurrency && exp.OriginalAmount > 0 {
			converted := map[string]interface{}{
				"type":   "text",
				"text":   fmt.Sprintf("≈ %s %s", formatAmount(exp.OriginalAmount), exp.OriginalCurrency),
				"size":   "xxs",
				"color":  "#64748B",
				"margin": "xs",
			}
			bodyContents = append(bodyContents, converted)
		}
	}

	body := map[string]interface{}{
		"type":       "box",
		"layout":     "vertical",
		"paddingAll": "16px",
		"contents":   bodyContents,
	}

	// Footer
	footer := map[string]interface{}{
		"type":       "box",
		"layout":     "vertical",
		"paddingAll": "12px",
		"contents": []interface{}{
			map[string]interface{}{
				"type":  "text",
				"text":  i18n.Tf(locale, "expense.count", map[string]string{"count": fmt.Sprintf("%d", len(expenses))}),
				"size":  "xxs",
				"color": "#64748B",
				"align": "center",
			},
		},
	}

	return map[string]interface{}{
		"type":   "bubble",
		"header": header,
		"body":   body,
		"footer": footer,
	}
}

func formatAmount(amount float64) string {
	if amount == float64(int64(amount)) {
		return fmt.Sprintf("%d", int64(amount))
	}
	return fmt.Sprintf("%.2f", amount)
}
```

**Step 4: Run tests**

Run: `go test ./internal/adapter/messenger/line/flex/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/adapter/messenger/line/flex/expense.go internal/adapter/messenger/line/flex/expense_test.go
git commit -m "feat(line/flex): add expense confirmation Flex Message builder"
```

---

### Task 7: Flex Message Builders — Report + Error Templates

**Files:**
- Create: `internal/adapter/messenger/line/flex/report.go`
- Create: `internal/adapter/messenger/line/flex/error.go`
- Create: `internal/adapter/messenger/line/flex/builders_test.go`

**Step 1: Write failing tests**

Create `internal/adapter/messenger/line/flex/builders_test.go`:

```go
package flex

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildReportBubble(t *testing.T) {
	bubble := BuildReportBubble("https://example.com/report/abc123", "zh-TW")

	require.NotNil(t, bubble)
	assert.Equal(t, "bubble", bubble["type"])
	assert.NotNil(t, bubble["header"])
	assert.NotNil(t, bubble["body"])
	assert.NotNil(t, bubble["footer"])
}

func TestBuildReportBubble_English(t *testing.T) {
	bubble := BuildReportBubble("https://example.com/report/abc123", "en")

	require.NotNil(t, bubble)
	assert.Equal(t, "bubble", bubble["type"])
}

func TestBuildErrorBubble(t *testing.T) {
	bubble := BuildErrorBubble("訊息中未偵測到任何支出", "請嘗試輸入如：「午餐 便當 85元」", "zh-TW")

	require.NotNil(t, bubble)
	assert.Equal(t, "bubble", bubble["type"])
	assert.NotNil(t, bubble["header"])
	assert.NotNil(t, bubble["body"])
}

func TestBuildInfoBubble(t *testing.T) {
	bubble := BuildInfoBubble("訊息中未偵測到任何支出", "請嘗試輸入如：「午餐 便當 85元」", "zh-TW")

	require.NotNil(t, bubble)
	assert.Equal(t, "bubble", bubble["type"])
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/adapter/messenger/line/flex/ -run "TestBuildReport|TestBuildError|TestBuildInfo" -v`
Expected: FAIL

**Step 3: Implement report builder**

Create `internal/adapter/messenger/line/flex/report.go`:

```go
package flex

import (
	"github.com/riverlin/aiexpense/internal/i18n"
)

// BuildReportBubble creates a LINE Flex Message bubble for report link.
func BuildReportBubble(reportURL, locale string) map[string]interface{} {
	header := map[string]interface{}{
		"type":            "box",
		"layout":          "vertical",
		"backgroundColor": "#4F46E5",
		"paddingAll":      "16px",
		"contents": []interface{}{
			map[string]interface{}{
				"type":  "text",
				"text":  i18n.T(locale, "flex.app_name"),
				"color": "#FFFFFF",
				"size":  "xs",
			},
			map[string]interface{}{
				"type":   "text",
				"text":   i18n.T(locale, "report.title"),
				"color":  "#FFFFFF",
				"size":   "xl",
				"weight": "bold",
				"margin": "sm",
			},
		},
	}

	body := map[string]interface{}{
		"type":       "box",
		"layout":     "vertical",
		"paddingAll": "16px",
		"contents": []interface{}{
			map[string]interface{}{
				"type":  "text",
				"text":  i18n.T(locale, "report.description"),
				"size":  "sm",
				"color": "#1E293B",
				"wrap":  true,
			},
			map[string]interface{}{
				"type":   "text",
				"text":   i18n.T(locale, "report.validity"),
				"size":   "xxs",
				"color":  "#64748B",
				"margin": "sm",
			},
		},
	}

	footer := map[string]interface{}{
		"type":       "box",
		"layout":     "vertical",
		"paddingAll": "12px",
		"contents": []interface{}{
			map[string]interface{}{
				"type":  "button",
				"style": "primary",
				"color": "#4F46E5",
				"action": map[string]interface{}{
					"type":  "uri",
					"label": i18n.T(locale, "report.button"),
					"uri":   reportURL,
				},
			},
		},
	}

	return map[string]interface{}{
		"type":   "bubble",
		"header": header,
		"body":   body,
		"footer": footer,
	}
}
```

**Step 4: Implement error/info builder**

Create `internal/adapter/messenger/line/flex/error.go`:

```go
package flex

import (
	"github.com/riverlin/aiexpense/internal/i18n"
)

// BuildErrorBubble creates a LINE Flex Message bubble for error messages.
func BuildErrorBubble(message, hint, locale string) map[string]interface{} {
	return buildMessageBubble(message, hint, locale, "#DC2626", i18n.T(locale, "error.title"))
}

// BuildInfoBubble creates a LINE Flex Message bubble for informational messages.
func BuildInfoBubble(message, hint, locale string) map[string]interface{} {
	return buildMessageBubble(message, hint, locale, "#64748B", i18n.T(locale, "flex.app_name"))
}

func buildMessageBubble(message, hint, locale, headerColor, title string) map[string]interface{} {
	header := map[string]interface{}{
		"type":            "box",
		"layout":          "vertical",
		"backgroundColor": headerColor,
		"paddingAll":      "16px",
		"contents": []interface{}{
			map[string]interface{}{
				"type":  "text",
				"text":  i18n.T(locale, "flex.app_name"),
				"color": "#FFFFFF",
				"size":  "xs",
			},
			map[string]interface{}{
				"type":   "text",
				"text":   title,
				"color":  "#FFFFFF",
				"size":   "xl",
				"weight": "bold",
				"margin": "sm",
			},
		},
	}

	bodyContents := []interface{}{
		map[string]interface{}{
			"type":  "text",
			"text":  message,
			"size":  "sm",
			"color": "#1E293B",
			"wrap":  true,
		},
	}

	if hint != "" {
		bodyContents = append(bodyContents, map[string]interface{}{
			"type":   "text",
			"text":   hint,
			"size":   "xxs",
			"color":  "#64748B",
			"margin": "md",
			"wrap":   true,
		})
	}

	body := map[string]interface{}{
		"type":       "box",
		"layout":     "vertical",
		"paddingAll": "16px",
		"contents":   bodyContents,
	}

	return map[string]interface{}{
		"type":   "bubble",
		"header": header,
		"body":   body,
	}
}
```

**Step 5: Run tests**

Run: `go test ./internal/adapter/messenger/line/flex/ -v`
Expected: PASS (all tests in both test files)

**Step 6: Commit**

```bash
git add internal/adapter/messenger/line/flex/report.go internal/adapter/messenger/line/flex/error.go internal/adapter/messenger/line/flex/builders_test.go
git commit -m "feat(line/flex): add report link and error/info Flex Message builders"
```

---

### Task 8: LINE Handler — Integrate Flex Messages + Locale

**Files:**
- Modify: `internal/adapter/messenger/line/handler.go`
- Modify: `internal/adapter/messenger/line/handler_test.go`

**Step 1: Update Handler to accept UserRepository and use Flex builders**

Modify `internal/adapter/messenger/line/handler.go`:

1. Add `userRepo domain.UserRepository` field to `Handler` struct
2. Update `NewHandler` to accept `userRepo domain.UserRepository`
3. In `HandleWebhook`, after `h.useCase.Execute()`, look up user locale via `h.userRepo.GetByID()`, then build the appropriate Flex Message based on `resp.Type`
4. Call `h.client.SendFlexReply()` instead of `h.client.SendReply()`
5. Keep `resp.Text` as the `altText` parameter for Flex Messages

The handler response section (currently lines 115-122) becomes:

```go
// Send reply as Flex Message
if resp != nil && h.client != nil {
    // Get user locale
    locale := "zh-TW"
    if h.userRepo != nil {
        if user, err := h.userRepo.GetByID(ctx, e.Source.UserID); err == nil && user.Locale != "" {
            locale = user.Locale
        }
    }

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
        if err := h.client.SendFlexReply(ctx, e.ReplyToken, resp.Text, flexBubble); err != nil {
            log.Printf("[LINE Webhook] Failed to send flex reply: %v", err)
            // Fallback to plain text
            if err := h.client.SendReply(ctx, e.ReplyToken, resp.Text); err != nil {
                log.Printf("[LINE Webhook] Fallback text reply also failed: %v", err)
            }
        }
    } else if resp.Text != "" {
        // Fallback for unknown types
        if err := h.client.SendReply(ctx, e.ReplyToken, resp.Text); err != nil {
            log.Printf("[LINE Webhook] Failed to send reply: %v", err)
        }
    }
}
```

Add helper functions in handler.go:

```go
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
```

**Step 2: Update handler tests**

In `internal/adapter/messenger/line/handler_test.go`, update `NewHandler` calls to include the new `userRepo` parameter (pass `nil`):

```go
handler := NewHandler("test_channel_secret", mockUC, nil, nil) // last nil = userRepo
```

**Step 3: Run tests**

Run: `go test ./internal/adapter/messenger/line/ -v`
Expected: PASS

Run: `go test ./internal/... -v -count=1`
Expected: PASS (full suite)

**Step 4: Commit**

```bash
git add internal/adapter/messenger/line/handler.go internal/adapter/messenger/line/handler_test.go
git commit -m "feat(line): integrate Flex Messages in LINE handler"
```

---

### Task 9: Auto-Signup — Fetch LINE Profile Language

**Files:**
- Modify: `internal/usecase/auto_signup.go`
- Modify: `internal/usecase/auto_signup_test.go`

**Step 1: Add ProfileFetcher interface and locale detection to auto_signup.go**

Add a `ProfileFetcher` interface and update `AutoSignupUseCase`:

```go
// ProfileFetcher fetches user profile data from a messenger platform
type ProfileFetcher interface {
    GetLanguage(ctx context.Context, userID string) (string, error)
}
```

Update `AutoSignupUseCase` struct to include `profileFetcher ProfileFetcher` (optional).

In `Execute()`, after creating the user, if `u.profileFetcher != nil` and `messengerType == "line"`:
1. Call `u.profileFetcher.GetLanguage(ctx, userID)`
2. Map the returned language to a supported locale (use `i18n.SupportedLocales()` to check)
3. If supported, update `user.Locale` via `u.userRepo.Update(ctx, user)`
4. If not supported, set `user.Locale` to `i18n.DefaultLocale()`

**Step 2: Update auto_signup_test.go**

Add test for locale detection:
- Mock ProfileFetcher that returns "zh-TW"
- Verify user is created with locale "zh-TW"
- Test fallback: mock returns "ja" (unsupported), verify user gets "zh-TW" default

**Step 3: Run tests**

Run: `go test ./internal/usecase/ -run TestAutoSignup -v`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/usecase/auto_signup.go internal/usecase/auto_signup_test.go
git commit -m "feat(signup): detect LINE user language on auto-signup"
```

---

### Task 10: Wire Everything Together in main.go

**Files:**
- Modify: `cmd/server/main.go`

**Step 1: Update LINE handler initialization**

Find where `line.NewHandler` is called in `cmd/server/main.go`. Update it to pass `userRepo` as the new parameter.

**Step 2: Create a LINE ProfileFetcher adapter**

Create a simple adapter that wraps `line.Client.GetProfile()` and implements the `ProfileFetcher` interface from Task 9:

```go
// In a new file or inline in main.go — a thin adapter
type lineProfileFetcher struct {
    client *line.Client
}

func (f *lineProfileFetcher) GetLanguage(ctx context.Context, userID string) (string, error) {
    profile, err := f.client.GetProfile(ctx, userID)
    if err != nil {
        return "", err
    }
    return profile.Language, nil
}
```

Pass it to `NewAutoSignupUseCase` when LINE is enabled.

**Step 3: Verify compilation and startup**

Run: `go build ./cmd/server/`
Expected: BUILD SUCCESS

**Step 4: Commit**

```bash
git add cmd/server/main.go
git commit -m "feat: wire LINE Flex Messages and i18n in server initialization"
```

---

### Task 11: Full Integration Test

**Files:**
- All modified/created files

**Step 1: Run full test suite**

Run: `go test ./... -v -count=1`
Expected: ALL PASS

**Step 2: Manual verification checklist**

- [ ] `go build ./cmd/server/` compiles cleanly
- [ ] `go vet ./...` reports no issues
- [ ] i18n tests pass with both zh-TW and en
- [ ] Flex builder tests produce valid bubble structures
- [ ] LINE handler tests pass with nil userRepo
- [ ] Profile API tests pass with mock HTTP server
- [ ] Auto-signup tests pass with locale detection
- [ ] Existing tests (all messengers) still pass

**Step 3: Final commit (if any fixes needed)**

```bash
git add -A
git commit -m "fix: resolve integration issues for LINE Flex Messages + i18n"
```

---

## Task Dependency Graph

```
Task 1 (translation files)
  └→ Task 2 (i18n engine) ──→ Task 6 (expense builder) ──→ Task 8 (handler integration)
                           ──→ Task 7 (report+error builders) ──↗
Task 3 (MessageResponse.Type) ──────────────────────────────────↗
Task 4 (SendFlexReply) ────────────────────────────────────────↗
Task 5 (Profile API) ──→ Task 9 (auto-signup locale) ──→ Task 10 (wiring)
                                                            ↓
                                                     Task 11 (integration test)
```

**Parallelizable groups:**
- Group A: Tasks 1→2 (i18n)
- Group B: Task 3 (domain), Task 4 (client), Task 5 (profile) — all independent
- Group C: Tasks 6+7 (flex builders, depend on Task 2)
- Group D: Task 8 (handler, depends on 3+4+6+7)
- Group E: Task 9 (depends on 2+5), Task 10 (depends on 8+9)
- Final: Task 11
