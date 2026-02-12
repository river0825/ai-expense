# LINE Flex Messages + i18n Design

**Date:** 2026-02-10
**Status:** Draft

## Goals

1. Replace plain text LINE bot responses with Flex Messages for a modern, styled experience
2. Add backend i18n support with `go:embed` JSON translation files, defaulting to `zh-TW`
3. Detect user language from LINE profile on signup, store in `User.Locale`

## Architecture

### New Components

```
internal/
  i18n/
    i18n.go              # T() and Tf() translation functions
    locales/
      zh-TW.json         # Traditional Chinese (default)
      en.json            # English
  adapter/messenger/line/
    flex/
      expense.go         # Expense confirmation Flex template
      report.go          # Report link Flex template
      error.go           # Error/info Flex template
    client.go            # Updated: add SendFlexReply method
    handler.go           # Updated: build Flex Messages, fetch LINE profile language
    profile.go           # NEW: LINE Profile API client
```

### Data Flow

```
User sends LINE message
  -> LINE Handler receives webhook
  -> ProcessMessageUseCase.Execute() returns MessageResponse{Type, Text, Data}
  -> Handler checks response.Type
  -> Flex builder creates Flex JSON using response.Data + user locale
  -> client.SendFlexReply(ctx, replyToken, altText, flexContainer)
```

### Non-LINE messengers

No changes. They continue reading `MessageResponse.Text` as plain text.

## Domain Changes

### MessageResponse (add Type field)

```go
type MessageResponse struct {
    Type string      `json:"type"`  // "expense", "report", "error", "info"
    Text string      `json:"text"`  // Plain text fallback (used by non-LINE messengers)
    Data interface{} `json:"data,omitempty"`
}
```

Response types:
- `"expense"` — Expense confirmation with structured data in `Data`
- `"report"` — Report link; `Data` contains `{"link": "https://..."}`
- `"error"` — Error message; `Text` contains the error string
- `"info"` — Informational message (e.g., "No expenses detected")

## i18n Package

### API

```go
package i18n

// T returns a translated string for the given locale and key.
// Falls back to zh-TW if key not found in requested locale.
func T(locale, key string) string

// Tf returns a translated string with named parameter substitution.
// Params like {count}, {amount} are replaced from the map.
func Tf(locale, key string, params map[string]string) string

// SupportedLocales returns all available locale codes.
func SupportedLocales() []string

// DefaultLocale returns "zh-TW".
func DefaultLocale() string
```

### Translation Keys

```json
{
  "expense.recorded": "已記錄 {count} 筆支出，合計：{amount} {currency}",
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
  "flex.total": "合計",
  "flex.date": "日期",
  "flex.category": "分類",
  "flex.account": "帳戶"
}
```

### Implementation

- Uses `go:embed` to load JSON files at compile time
- Parsed once at init into `map[string]map[string]string` (locale -> key -> value)
- `Tf` replaces `{paramName}` patterns using `strings.NewReplacer`
- Falls back: requested locale -> zh-TW -> raw key

## Flex Message Templates

### Design System

| Element | Color | Hex |
|---------|-------|-----|
| Header BG (success) | Deep Emerald | `#059669` |
| Header BG (report) | Rich Indigo | `#4F46E5` |
| Header BG (error) | Warm Red | `#DC2626` |
| Body BG | Clean White | `#FFFFFF` |
| Primary Text | Near Black | `#1E293B` |
| Muted Text | Slate | `#64748B` |
| Amount Text | Deep Emerald | `#059669` |
| CTA Button | Indigo | `#4F46E5` |

### Template 1: Expense Confirmation

Bubble with:
- **Header** (`#059669`): App name "AIExpense", summary "已記錄 3 筆支出", total "合計 NT$1,250" in bold white
- **Body** (white): Each expense as a row:
  - Description + amount (horizontal, bold)
  - Category + date + account (muted, small)
  - Separator between items
- **Footer**: "{count} 筆支出已成功記錄" in muted text

```
┌──────────────────────────────┐
│  ███ #059669 ████████████   │
│  AIExpense                   │
│  已記錄 3 筆支出             │
│  合計 NT$1,250               │
├──────────────────────────────┤
│  午餐 便當               $85 │
│  餐飲 · 2024-02-10 · 現金   │
│  ─────────────────────────── │
│  高鐵票 台北→高雄     $1,065 │
│  交通 · 2024-02-10 · 信用卡 │
│  ─────────────────────────── │
│  咖啡                   $100 │
│  餐飲 · 2024-02-10 · 現金   │
├──────────────────────────────┤
│  3 筆支出已成功記錄          │
└──────────────────────────────┘
```

### Template 2: Report Link

Bubble with:
- **Header** (`#4F46E5`): App name, "支出報告"
- **Body**: Description text + validity note (muted)
- **Footer**: Indigo URI button "查看報告" linking to the report URL

```
┌──────────────────────────────┐
│  ███ #4F46E5 ████████████   │
│  AIExpense                   │
│  支出報告                    │
├──────────────────────────────┤
│  您的個人支出報告已準備好    │
│  連結有效期 5 分鐘           │
├──────────────────────────────┤
│  ┌────────────────────────┐  │
│  │      查看報告          │  │
│  └────────────────────────┘  │
└──────────────────────────────┘
```

### Template 3: Error / Info

Bubble with:
- **Header** (`#DC2626` for errors, `#64748B` for info): App name, title
- **Body**: Error/info message text, optional hint in muted text

```
┌──────────────────────────────┐
│  ███ #DC2626 ████████████   │
│  AIExpense                   │
│  處理失敗                    │
├──────────────────────────────┤
│  訊息中未偵測到任何支出      │
│  請嘗試輸入如：              │
│  「午餐 便當 85元」          │
└──────────────────────────────┘
```

## LINE Profile Language Fetch

### New: `profile.go`

```go
// GetProfile fetches the user's LINE profile including language.
// GET https://api.line.me/v2/bot/profile/{userId}
// Returns: displayName, language, pictureUrl, statusMessage
func (c *Client) GetProfile(ctx context.Context, userID string) (*LineProfile, error)

type LineProfile struct {
    DisplayName   string `json:"displayName"`
    Language      string `json:"language"`      // e.g., "zh-TW", "en", "ja"
    PictureURL    string `json:"pictureUrl"`
    StatusMessage string `json:"statusMessage"`
}
```

### Integration with Auto-Signup

During `AutoSignupUseCase.Execute()`:
- If user is new (first message), fetch LINE profile
- Map LINE's `language` field to a supported locale (zh-TW, en)
- Store in `User.Locale`
- If language not in supported locales, default to `zh-TW`

## LINE Client Changes

### New method: `SendFlexReply`

```go
func (c *Client) SendFlexReply(ctx context.Context, replyToken, altText string, flex interface{}) error
```

The reply message payload changes from:
```json
{"replyToken": "...", "messages": [{"type": "text", "text": "..."}]}
```
To:
```json
{
  "replyToken": "...",
  "messages": [{
    "type": "flex",
    "altText": "已記錄 3 筆支出，合計 NT$1,250",
    "contents": { ... flex bubble JSON ... }
  }]
}
```

`altText` is the plain text fallback shown in notifications and devices that don't support Flex Messages. We use `MessageResponse.Text` for this.

## ProcessMessageUseCase Changes

Set `Type` field on all returned `MessageResponse`:

| Scenario | Type | Data |
|----------|------|------|
| Expenses recorded | `"expense"` | `[]map[string]interface{}` (existing) |
| Report link generated | `"report"` | `map[string]string{"link": "..."}` |
| No expenses detected | `"info"` | `nil` |
| Signup error | `"error"` | `nil` |
| Parse error | `"error"` | `nil` |

## Files Changed

| File | Change |
|------|--------|
| `internal/domain/messenger.go` | Add `Type` field to `MessageResponse` |
| `internal/usecase/process_message.go` | Set `Type` on all responses |
| `internal/i18n/i18n.go` | New: translation engine |
| `internal/i18n/locales/zh-TW.json` | New: Chinese translations |
| `internal/i18n/locales/en.json` | New: English translations |
| `internal/adapter/messenger/line/flex/expense.go` | New: expense Flex builder |
| `internal/adapter/messenger/line/flex/report.go` | New: report Flex builder |
| `internal/adapter/messenger/line/flex/error.go` | New: error/info Flex builder |
| `internal/adapter/messenger/line/profile.go` | New: LINE profile API |
| `internal/adapter/messenger/line/client.go` | Add `SendFlexReply` method |
| `internal/adapter/messenger/line/handler.go` | Use Flex builders + locale |
| `internal/usecase/auto_signup.go` | Fetch LINE profile language on signup |
| Tests for all new code | |

## Out of Scope

- Rich messages for non-LINE messengers (Slack Block Kit, Telegram HTML, etc.)
- `/lang` command for manual locale override
- Additional locales beyond zh-TW and en
