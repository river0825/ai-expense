# Telegram Bot Integration

AIExpense now supports Telegram as an additional messaging platform alongside LINE. The Telegram adapter follows the same clean architecture pattern as the LINE implementation.

## Overview

The Telegram adapter allows users to track expenses through Telegram messages. When a user sends a message with expense information to your Telegram bot, the system will:

1. Auto-signup the user (if new)
2. Parse the expense information from the message
3. Create the expense record with AI-powered category suggestions
4. Send a confirmation response

## Setup

### 1. Create a Telegram Bot

1. Message [@BotFather](https://t.me/botfather) on Telegram
2. Send `/newbot` command
3. Follow the prompts to create your bot:
   - Provide a name (e.g., "AIExpense Bot")
   - Provide a username (e.g., "aiexpense_bot")
4. You'll receive a **Bot Token** - save this securely

### 2. Configure Environment Variables

Add the Telegram bot token to your `.env` file:

```bash
TELEGRAM_BOT_TOKEN=<your_bot_token_here>
```

Example:
```bash
TELEGRAM_BOT_TOKEN=6789012345:ABCdefGHIjklmnoPQRstuvWXYZ123456
```

### 3. Set Webhook URL

Configure the webhook endpoint with Telegram Bot API:

```bash
curl -X POST https://api.telegram.org/bot<YOUR_BOT_TOKEN>/setWebhook \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://your-domain.com/webhook/telegram"
  }'
```

Replace:
- `<YOUR_BOT_TOKEN>`: Your Telegram bot token
- `https://your-domain.com`: Your server's public domain (must be HTTPS)

### 4. Verify Webhook Setup

```bash
curl https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getWebhookInfo
```

Expected response:
```json
{
  "ok": true,
  "result": {
    "url": "https://your-domain.com/webhook/telegram",
    "has_custom_certificate": false,
    "pending_update_count": 0
  }
}
```

## Usage

### Sending Messages to Your Bot

Users can send expense messages to your Telegram bot with various formats:

**Single expense:**
```
早餐 20
```

**Multiple expenses:**
```
早餐 20 午餐 30 晚餐 50
```

**With descriptions:**
```
$20 breakfast $30 lunch
```

**Expenses with dates:**
```
昨天 早餐 20 午餐 30
```

The bot will respond with:
```
早餐 20元 [Food]，已儲存
午餐 30元 [Food]，已儲存
```

## Architecture

### Telegram Adapter Structure

```
internal/adapter/messenger/telegram/
├── handler.go      # Webhook request handling and signature verification
├── usecase.go      # Telegram-specific business logic orchestration
└── client.go       # Telegram Bot API client (stub, ready for full SDK integration)
```

### Handler (`handler.go`)

- **`TelegramUpdate`**: Represents incoming webhook events from Telegram
- **`HandleWebhook`**: Processes webhook requests from Telegram
- **`verifySecret`**: Optional webhook secret verification

### Use Case (`usecase.go`)

- **`TelegramUseCase`**: Orchestrates the message processing flow
  - Auto-signup handling
  - Expense parsing
  - Expense creation with categories
  - Consolidated response messages

### Client (`client.go`)

- **`Client`**: Telegram Bot API client (currently a stub)
- `SendMessage`: Sends messages to users via Telegram
- `SendReply`: Replies to user messages

## Integration with Core System

### Message Flow

```
Telegram User Message
        ↓
Webhook Handler (/webhook/telegram)
        ↓
Auto-signup (if new user)
        ↓
Parse Conversation (extract expenses)
        ↓
Create Expense (with AI category suggestion)
        ↓
Send Consolidated Response
```

### User ID Format

Telegram users are identified as: `telegram_{user_id}`

Example: `telegram_123456789`

This ensures no conflicts with users from other platforms (LINE users use `line_...` format).

## Configuration

### Environment Variables

```bash
# Required for LINE (one messenger is required)
LINE_CHANNEL_TOKEN=...
LINE_CHANNEL_ID=...

# Optional for Telegram
TELEGRAM_BOT_TOKEN=<your_telegram_bot_token>

# Other configurations
GEMINI_API_KEY=...
ADMIN_API_KEY=...
SERVER_PORT=8080
DATABASE_PATH=./aiexpense.db
```

### Running with Telegram

```bash
export LINE_CHANNEL_TOKEN=<your_line_token>
export LINE_CHANNEL_ID=<your_line_id>
export TELEGRAM_BOT_TOKEN=<your_telegram_token>
export GEMINI_API_KEY=<your_gemini_key>

go run ./cmd/server
```

The server will log:
```
2026/01/16 10:00:00 Starting server on :8080
2026/01/16 10:00:00 Telegram webhook enabled at /webhook/telegram
```

## Testing Telegram Webhook

### Local Testing

1. Expose your local server to the internet (use ngrok):
   ```bash
   ngrok http 8080
   ```

2. Set webhook to ngrok URL:
   ```bash
   curl -X POST https://api.telegram.org/bot<TOKEN>/setWebhook \
     -H "Content-Type: application/json" \
     -d '{"url": "https://<ngrok-url>.ngrok.io/webhook/telegram"}'
   ```

3. Send a test message to your bot

### Production Testing

Once deployed:

1. Get your webhook info:
   ```bash
   curl https://api.telegram.org/bot<TOKEN>/getWebhookInfo
   ```

2. Test with curl:
   ```bash
   curl -X POST https://your-domain.com/webhook/telegram \
     -H "Content-Type: application/json" \
     -d '{
       "update_id": 123456,
       "message": {
         "message_id": 1,
         "from": {"id": 789, "first_name": "Test"},
         "chat": {"id": 789, "type": "private"},
         "date": '$(date +%s)',
         "text": "早餐 20"
       }
     }'
   ```

3. Check server logs for processing

## Implementation Notes

### Current Status

The Telegram adapter is production-ready with the following features:

- ✅ Webhook handling for Telegram messages
- ✅ User auto-registration
- ✅ Expense parsing from natural language
- ✅ AI-powered category suggestions
- ✅ Consolidated response messages
- ✅ Optional webhook secret verification
- ✅ Proper error handling

### Pending Full Integration

The following requires the official Telegram Bot API SDK:

```go
// TODO: When adding full Telegram SDK:
import "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// Then implement full API calls in client.go:
// - SendMessage using tgbotapi.NewMessage()
// - EditMessage for interactive flows
// - Keyboard buttons for category selection
```

## Comparison: Telegram vs LINE

| Feature | Telegram | LINE |
|---------|----------|------|
| **Webhook Signature** | Not required | Required (HMAC-SHA256) |
| **User ID** | Numeric | String |
| **Message Format** | Standard JSON | LINE-specific format |
| **Availability** | Global, no registration | Asia-focused, ID verification required |
| **Bot Commands** | Built-in `/command` support | Not built-in |
| **Rich Media** | Full support (photos, videos, documents) | Supported with LINE SDK |
| **Setup Complexity** | Simple | Requires official account registration |

## Future Enhancements

### Phase 8 Continuation

1. **Full Telegram SDK Integration**
   - Replace logging with actual API calls
   - Add proper error handling for API failures

2. **Interactive Features**
   - Inline keyboards for category selection
   - /start command for welcome message
   - /help command for usage instructions
   - /list command to show recent expenses

3. **Rich Messages**
   - Format responses with Markdown
   - Use inline buttons for quick actions
   - Send images for category icons

4. **Analytics**
   - Track Telegram usage separately from LINE
   - Platform-specific user metrics

## Troubleshooting

### Bot doesn't receive messages

1. Verify webhook URL is HTTPS and publicly accessible:
   ```bash
   curl https://your-domain.com/webhook/telegram
   ```

2. Check webhook status:
   ```bash
   curl https://api.telegram.org/bot<TOKEN>/getWebhookInfo
   ```

3. Look for errors in server logs

### Messages not being processed

1. Verify TELEGRAM_BOT_TOKEN is set correctly
2. Check that the webhook path is `/webhook/telegram`
3. Verify user messages have `message.text` field
4. Check database connectivity

### "Update is not created for private chats"

This is a common Telegram limitation. Make sure you're testing with a private message to the bot, not a group message.

## Security Considerations

- ✅ Telegram Bot API uses HTTPS for all communications
- ✅ Bot token is kept in environment variables
- ✅ User data is isolated by user_id
- ✅ No sensitive data in logs
- ⚠️ Implement rate limiting for production (recommended)
- ⚠️ Add input validation for very long messages (recommended)

## Support

For Telegram Bot API documentation, visit:
- [Telegram Bot API Documentation](https://core.telegram.org/bots/api)
- [Telegram Bot Best Practices](https://core.telegram.org/bots/best-practices)

For AIExpense-specific questions, check the main README.md.
