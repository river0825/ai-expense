# Slack Bot Integration Guide

AIExpense now supports Slack, allowing users to track expenses directly through Slack messages and app mentions.

## Table of Contents

1. [Setup Instructions](#setup-instructions)
2. [Architecture](#architecture)
3. [Features](#features)
4. [Configuration](#configuration)
5. [Webhook Setup](#webhook-setup)
6. [API Reference](#api-reference)
7. [Message Flow](#message-flow)
8. [Troubleshooting](#troubleshooting)

## Setup Instructions

### Prerequisites

- A Slack workspace (with appropriate permissions to create apps)
- AIExpense server running and accessible via HTTPS
- Slack app credentials (bot token and signing secret)

### Step 1: Create a Slack App

1. Go to https://api.slack.com/apps
2. Click "Create New App"
3. Choose "From scratch"
4. **App Name**: "AIExpense" (or your preferred name)
5. **Select Workspace**: Choose your Slack workspace
6. Click "Create App"

### Step 2: Configure Bot User

1. From the left sidebar, click **"App Home"**
2. Under "Your App's Presence", click **"Edit"**
3. Toggle on:
   - ✅ Always show my bot as online
   - ✅ Show tabs for home and messages

### Step 3: Obtain Credentials

1. Click **"Install App"** in the left sidebar
2. Click **"Install to Workspace"** and authorize
3. Copy the **Bot User OAuth Token** (starts with `xoxb-`)
4. Go to **"Basic Information"** tab
5. Scroll to **"App Credentials"**
6. Copy the **Signing Secret**

### Step 4: Enable Events Subscriptions

1. From the left sidebar, click **"Event Subscriptions"**
2. Toggle **"Enable Events"** to ON
3. You'll be asked for a **Request URL** - set this to: `https://your-domain.com/webhook/slack`
4. Wait for verification (it will make a POST request and expect a 200 response with challenge)
5. Under **"Subscribe to bot events"**, add these events:
   - `message.im` - Direct messages
   - `app_mention` - When the app is mentioned
6. Click **"Save Changes"**

### Step 5: Set Permissions (Scopes)

1. Click **"OAuth & Permissions"** in the left sidebar
2. Scroll to **"Scopes"** under **"Bot Token Scopes"**
3. Add these scopes:
   - `chat:write` - Send messages
   - `channels:read` - Read channel info
   - `im:read` - Read direct messages
   - `users:read` - Read user profiles
4. Save changes

### Step 6: Configure Environment Variables

Add these to your `.env` file:

```bash
SLACK_BOT_TOKEN=xoxb-your-token-here
SLACK_SIGNING_SECRET=your-signing-secret-here
```

### Step 7: Start the Server

```bash
export SLACK_BOT_TOKEN=xoxb-your-token-here
export SLACK_SIGNING_SECRET=your-signing-secret-here
go run ./cmd/server
```

## Architecture

The Slack integration follows the same adapter pattern as other messengers (LINE, Telegram, Discord, WhatsApp):

```
Slack User Message
        ↓
Webhook Handler (verify signature)
        ↓
Auto-signup (if new user)
        ↓
Parse Conversation (extract expenses)
        ↓
Create Expense (with AI categorization)
        ↓
Send Response (via Slack API)
```

### File Structure

```
internal/adapter/messenger/slack/
├── client.go      # Slack Bot API HTTP client
├── handler.go     # Webhook request handler
└── usecase.go     # Business logic orchestration
```

## Features

### 1. Direct Messages

Send expense information directly to the bot:

```
User: breakfast $8 lunch $12
Bot: ✅ breakfast - $8.00 (Food)
     ✅ lunch - $12.00 (Food)
     Recorded 2 expense(s)
```

### 2. App Mentions in Channels

Mention the bot in a channel:

```
User: @AIExpense spent $50 on groceries
Bot: ✅ spent on groceries - $50.00 (Shopping)
     Recorded 1 expense(s)
```

### 3. Multi-Expense Support

Record multiple expenses in one message:

```
User: coffee $5, lunch $12, dinner $25
Bot: ✅ coffee - $5.00 (Food)
     ✅ lunch - $12.00 (Food)
     ✅ dinner - $25.00 (Food)
     Recorded 3 expense(s)
```

### 4. AI-Powered Categorization

Expenses are automatically categorized using AI:
- Food & Dining
- Transportation
- Shopping
- Entertainment
- Utilities
- Other

### 5. Error Handling

When parsing fails:

```
User: invalid data
Bot: I didn't find any expenses to record. Try saying something like:
     • 'breakfast $8'
     • 'lunch 12 coffee 5'
     • 'spent $50 on groceries'
```

## Configuration

### Environment Variables

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `SLACK_BOT_TOKEN` | Yes | Bot user OAuth token | `xoxb-123456...` |
| `SLACK_SIGNING_SECRET` | No* | Signing secret for webhook verification | `abc123def456...` |

*While optional for development, **strongly recommended for production**

### Optional Configuration

To support multiple Slack workspaces, you can run multiple instances with different tokens.

## Webhook Setup

### Request Verification

The handler automatically verifies all Slack requests using HMAC-SHA256:

1. **Timestamp Check**: Requests older than 5 minutes are rejected
2. **Signature Verification**: Compares `X-Slack-Request-Signature` header with computed HMAC
3. **Challenge Response**: Automatically responds to URL verification challenges

### Webhook Events

The handler processes these event types:

#### Message Events
- **Type**: `message`
- **Condition**: Direct message (not in channel unless bot is mentioned)
- **Processing**: Automatic expense parsing and creation

#### App Mention Events
- **Type**: `app_mention`
- **Condition**: Bot is mentioned (`@AIExpense`) anywhere
- **Processing**: Extracts message text and processes for expenses

#### Ignored Events
- Bot's own messages (filtered by `bot_id`)
- Messages without text
- Messages without user ID
- Thread replies (can be enabled if needed)

## API Reference

### Client Methods

#### SendMessage

```go
func (c *Client) SendMessage(userID, text string) error
```

Sends a message to a user or channel.

**Parameters:**
- `userID`: Slack user ID (e.g., `U12345678`)
- `text`: Message text (supports markdown)

**Example:**
```go
client.SendMessage("U12345678", "✅ Lunch recorded - $12.00")
```

#### GetBotInfo

```go
func (c *Client) GetBotInfo() (map[string]interface{}, error)
```

Retrieves bot authentication info (used for testing).

**Returns:**
```json
{
  "ok": true,
  "url": "https://example.slack.com/",
  "team": "Example Team",
  "user": "aiexpense-bot",
  "team_id": "T12345678",
  "user_id": "U87654321",
  "bot_id": "B12345678"
}
```

#### OpenConversation

```go
func (c *Client) OpenConversation(userID string) (string, error)
```

Opens a direct message conversation with a user.

**Parameters:**
- `userID`: Slack user ID

**Returns:**
- `string`: Channel ID for the conversation
- `error`: If operation fails

### UseCase Methods

#### ProcessMessage

```go
func (u *UseCase) ProcessMessage(ctx context.Context, userID, text string) error
```

Processes an incoming message for expense parsing and creation.

**Flow:**
1. Auto-signup user (if new)
2. Parse text for expenses
3. Create expense records
4. Send confirmation

#### ProcessAppMention

```go
func (u *UseCase) ProcessAppMention(ctx context.Context, userID, text string) error
```

Processes a message where the bot is mentioned.

**Flow:**
1. Remove bot mention from text
2. Call ProcessMessage with cleaned text

## Message Flow

### Complete Flow Diagram

```
┌─────────────────┐
│  Slack User     │
│  sends message  │
└────────┬────────┘
         │
         v
┌─────────────────────────────┐
│  POST /webhook/slack        │
│  (Slack sends event)        │
└────────┬────────────────────┘
         │
         v
┌─────────────────────────────┐
│  Verify Signature           │
│  (HMAC-SHA256)              │
└────────┬────────────────────┘
         │
         v
┌─────────────────────────────┐
│  Parse Event                │
│  (Message, AppMention, etc) │
└────────┬────────────────────┘
         │
         v
┌─────────────────────────────┐
│  Auto-Signup                │
│  (if new user)              │
└────────┬────────────────────┘
         │
         v
┌─────────────────────────────┐
│  Parse Conversation         │
│  (Extract expenses from     │
│   natural language)         │
└────────┬────────────────────┘
         │
         v
┌─────────────────────────────┐
│  Create Expenses            │
│  (Save to database)         │
└────────┬────────────────────┘
         │
         v
┌─────────────────────────────┐
│  Send Response              │
│  (Confirmation message      │
│   via Slack API)            │
└────────┬────────────────────┘
         │
         v
┌─────────────────┐
│  Slack User     │
│  sees response  │
└─────────────────┘
```

### Example Conversation

**User Message:**
```
breakfast $8 lunch $12
```

**Internal Processing:**
```
1. Verify webhook signature ✓
2. Parse event (type: message, user: U12345678)
3. Auto-signup: slack_U12345678 ✓
4. Parse conversation:
   - breakfast: $8
   - lunch: $12
5. Create expenses:
   - breakfast → Food category → ID: exp_001
   - lunch → Food category → ID: exp_002
6. Send message via Slack API
```

**Bot Response:**
```
✅ breakfast - $8.00 (Food)
✅ lunch - $12.00 (Food)
Recorded 2 expense(s)
```

## Troubleshooting

### Webhook Not Verified

**Error**: "Failed to verify your request URL. It may have taken longer than 3 seconds to respond."

**Solution:**
1. Ensure your server is publicly accessible via HTTPS
2. Check that `/webhook/slack` endpoint returns 200 OK with challenge
3. Verify endpoint is `POST /webhook/slack`
4. Check server logs for errors

### Signature Verification Failed

**Error**: "signature verification failed"

**Solution:**
1. Verify `SLACK_SIGNING_SECRET` environment variable matches your app
2. Ensure your server's clock is synchronized (webhook timestamp validation)
3. Check that all headers are being received (`X-Slack-Request-Signature`, `X-Slack-Request-Timestamp`)

### Messages Not Being Received

**Symptoms**: No response from bot when sending messages

**Solutions:**
1. Verify bot has Direct Message permissions (is a workspace member)
2. Check event subscriptions: `message.im` and `app_mention` should be enabled
3. Verify bot has correct scopes: `chat:write`, `im:read`, `app_mentions:read`
4. Check server logs for webhook delivery errors
5. Test with Slack's Event Subscriptions debug panel

### Bot Not Responding

**Symptoms**: Bot receives events but doesn't send responses

**Possible causes:**
1. `SLACK_BOT_TOKEN` is invalid or expired
2. Bot doesn't have `chat:write` permission
3. User is not in the workspace member directory
4. Parsing errors (check server logs)

**Debug:**
```bash
# Check bot token validity
curl -X POST https://slack.com/api/auth.test \
  -H "Authorization: Bearer $SLACK_BOT_TOKEN"
```

### Database Errors

**Symptoms**: Expenses are parsed but not created

**Solutions:**
1. Verify SQLite database file exists and is writable
2. Check `DATABASE_PATH` environment variable
3. Run migrations: database should auto-initialize on startup
4. Check server logs for SQL errors

### Messages Cut Off or Formatted Incorrectly

**Solutions:**
1. Slack messages support markdown formatting
2. Use `*bold*` for emphasis, `_italic_` for italics
3. Use `\n` for line breaks in API calls
4. Slack automatically formats URLs and mentions
5. Check message length (Slack has a ~4000 character limit per message)

## Best Practices

### 1. Message Formatting

For better readability, use:

```
✅ Expense recorded
📊 Summary: 5 expenses, $125.50 total
⚠️ Error message
```

### 2. Error Recovery

The bot automatically:
- Continues processing after individual expense failures
- Provides summary of successes and failures
- Suggests proper format if no expenses are found

### 3. Security

- Always use HTTPS for webhook endpoints
- Verify signing secret in production
- Rotate tokens regularly
- Use environment variables, never hardcode credentials
- Limit bot permissions to minimum required (chat:write, im:read)

### 4. User Experience

- Provide immediate feedback (within 3 seconds)
- Use emoji for quick status indication
- Include totals and counts in responses
- Suggest format if user input is ambiguous

## Advanced Usage

### Direct User ID Format

User IDs from Slack are prefixed with `slack_` in the system:

```
Slack User ID: U12345678
AIExpense User ID: slack_U12345678
```

This allows the same user to have separate expense histories across different messengers.

### Processing Recurring Expenses

The system supports recurring expenses, which can be created through the REST API:

```bash
curl -X POST http://localhost:8080/api/recurring \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "slack_U12345678",
    "description": "Slack subscription",
    "amount": 99.99,
    "frequency": "monthly",
    "start_date": "2024-01-01T00:00:00Z"
  }'
```

### Analytics via REST API

After tracking expenses through Slack, access analytics:

```bash
# Get monthly report
curl -X POST http://localhost:8080/api/reports/generate \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "slack_U12345678",
    "period": "monthly",
    "include_breakdown": true
  }'
```

## Performance Considerations

- **Async Processing**: Webhook responses are sent asynchronously to meet Slack's 3-second timeout
- **Rate Limiting**: Slack enforces rate limits; built-in handling with exponential backoff
- **Webhook Timeout**: If processing takes >3 seconds, response is sent asynchronously
- **Concurrent Users**: Designed for horizontal scaling with stateless webhook handlers

## Support & Debugging

### Enable Debug Logging

Add to environment:
```bash
LOG_LEVEL=debug
```

### Slack Event Debugger

Use Slack's built-in event debugger:
1. Go to Event Subscriptions
2. Scroll to "Recent Events"
3. View event payload and response details

### Test Webhook Locally

```bash
# Install ngrok to create tunnel
ngrok http 8080

# Use ngrok URL in Slack webhook URL
# https://xxxxx.ngrok.io/webhook/slack
```

## Limitations & Future Enhancements

### Current Limitations

- Thread support: Can be enabled in future versions
- File uploads: Not yet supported
- Slash commands: Can be added in future versions
- Modal dialogs: Can be added for more complex interactions

### Planned Features

- [ ] Slash command `/expense` for structured input
- [ ] Modal dialogs for complex expense entry
- [ ] Scheduled expense reminders
- [ ] Budget alerts via direct message
- [ ] Expense history quick lookups
- [ ] Share expenses with team members
- [ ] Monthly report generation in channel

## Contributing

Found a bug or have a feature request? Open an issue in the repository.

---

**Last Updated**: January 2026
**Version**: 1.0.0
**Status**: Production Ready ✅
