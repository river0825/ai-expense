# AI-Powered Intent Detection Design

**Date**: 2026-03-05
**Issue**: #11
**Branch**: `11-feat-ai-powered-intent-detection-with-regex-first-llm-fallback`
**Status**: Implementation

## Problem

The chatbot only responds to explicit keyword commands (e.g., "切換 JPY", "report"). Contextual messages like "我正在日本旅行" (I'm traveling to Japan) are sent to the expense parser, which returns "No expenses detected." The bot should understand conversational context and proactively suggest actions.

## Solution: Hybrid Pipeline (Regex-first, LLM Fallback)

```
User Message
  │
  ├─ 1. Pending conversation state? → handle confirmation/follow-up
  ├─ 2. Keyword: currency change (isSetCurrencyIntent) → direct currency change
  ├─ 3. Keyword: report (isReportIntent) → generate report link
  ├─ 4. **NEW: LLM intent classification** → classify & act
  │     ├─ TRAVEL_CONTEXT → save pending state, suggest currency switch
  │     ├─ CURRENCY_CHANGE → handle edge cases keywords missed
  │     ├─ ADD_EXPENSE → fall through to expense parser
  │     └─ UNKNOWN → fall through to expense parser
  ├─ 5. Idempotency check
  └─ 6. Expense parsing via Gemini (existing)
```

### Why Regex-first?
- Keywords are free (no API call, no latency)
- LLM only fires when keywords fail — cost-effective
- Existing keyword behavior is preserved exactly

## Tier 1 Intents

| Intent | Trigger Example | Action |
|---|---|---|
| `TRAVEL_CONTEXT` | "我正在日本旅行", "I'm in Tokyo" | Suggest currency switch → confirmation flow |
| `CURRENCY_CHANGE` | Edge cases keywords miss | Direct currency change |
| `ADD_EXPENSE` | "午餐 300", "coffee $5" | Pass to existing expense parser |
| `REPORT` | Edge cases keywords miss | Generate report link |
| `UNKNOWN` | Anything unclassified | Pass to existing expense parser |

## Domain Types

```go
// IntentType represents the classified intent of a user message
type IntentType string

const (
    IntentAddExpense     IntentType = "ADD_EXPENSE"
    IntentTravelContext  IntentType = "TRAVEL_CONTEXT"
    IntentCurrencyChange IntentType = "CURRENCY_CHANGE"
    IntentReport         IntentType = "REPORT"
    IntentUnknown        IntentType = "UNKNOWN"
)

// ClassifiedIntent represents the result of AI intent classification
type ClassifiedIntent struct {
    Type                IntentType
    Confidence          float64
    Parameters          map[string]string // e.g., {"destination": "Japan", "currency": "JPY"}
    NeedsConfirmation   bool
    ConfirmationMessage string
}
```

## AI Service Extension

```go
// Added to ai.Service interface:
ClassifyIntent(ctx context.Context, text string, userCtx *domain.UserContext) (*ClassifyIntentResponse, error)

// ClassifyIntentResponse wraps classified intent with token metadata
type ClassifyIntentResponse struct {
    Intent  *domain.ClassifiedIntent
    Tokens  *TokenMetadata
    RawResponse string
}
```

## Confirmation Flow

When the LLM detects `TRAVEL_CONTEXT`:

1. Bot saves `ConversationState` with `ActiveIntent: "suggestion.currency.change"` and `PendingSlots: {"suggested_currency": "JPY", "destination": "Japan"}`
2. Bot replies: "你正在日本旅行！要切換預設幣別為 JPY 嗎？" / "You're in Japan! Switch default currency to JPY?"
3. User confirms with "OK", "好", "yes", "對" → update `DefaultInputCurrency` to JPY
4. User cancels with "cancel", "取消", "no" → clear pending state

## Gemini Prompt Design

```
You are an intent classifier for an expense tracking chatbot.
Classify the user's message into one of: TRAVEL_CONTEXT, CURRENCY_CHANGE, ADD_EXPENSE, REPORT, UNKNOWN.

Return JSON:
{
  "intent": "TRAVEL_CONTEXT",
  "confidence": 0.95,
  "parameters": {"destination": "Japan", "currency": "JPY"},
  "needs_confirmation": true,
  "confirmation_message_zh": "你正在日本旅行！要切換預設幣別為 JPY 嗎？",
  "confirmation_message_en": "You're in Japan! Switch default currency to JPY?"
}
```

## Insertion Point

In `process_message.go Execute()`, after `isReportIntent` check (line ~177) and before idempotency check (line ~179):

```go
// NEW: LLM intent classification (only if keywords didn't match)
if resp, handled := u.handleAIIntentDetection(ctx, msg, msgLower); handled {
    botReply = resp.Text
    return resp, nil
}
```

## Dependencies

- `ProcessMessageUseCase` needs access to `ai.Service` (new field: `aiService`)
- Existing `ConversationState` / `ConversationStateRepository` reused for confirmation flow
- No new database migrations needed — reuses existing `conversation_states` table

## Testing Strategy (TDD)

1. Travel context → LLM classifies as TRAVEL_CONTEXT → saves pending state → returns suggestion
2. Confirmation "OK"/"好" on pending `suggestion.currency.change` → updates DefaultInputCurrency
3. Confirmation "cancel"/"取消" on pending suggestion → clears state
4. Expired suggestion state → asks to restate
5. Regression: keyword currency intent still works without LLM call
6. Regression: expense message still works without LLM call
7. LLM returns ADD_EXPENSE → falls through to expense parser
8. LLM returns UNKNOWN → falls through to expense parser
9. LLM error → graceful fallback to expense parser
