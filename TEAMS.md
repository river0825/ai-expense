# Microsoft Teams Bot Integration Guide

AIExpense now supports Microsoft Teams, allowing users to track expenses directly through Teams messages, direct chats, and channel mentions.

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

- Microsoft Azure subscription
- Azure Bot Service credentials
- AIExpense server running and accessible via HTTPS
- Teams app credentials (App ID and App Password)

### Step 1: Register Bot with Azure Bot Service

1. Go to [Azure Portal](https://portal.azure.com)
2. Create a new resource: **Azure Bot**
3. **Resource name**: "AIExpense" (or your preferred name)
4. **Bot handle**: Choose a unique name (this becomes your bot's username)
5. **Resource group**: Create or select existing
6. **Pricing tier**: Select Free tier for development
7. Click **Create**

### Step 2: Obtain Bot Credentials

After creating the bot:

1. Go to **Settings** → **Configuration**
2. Copy the **Microsoft App ID**
3. Click **Create New Password** under Microsoft App password
4. Copy the generated **Password** (this won't be shown again)

Store these securely:
- **App ID**: Used as `TEAMS_APP_ID`
- **App Password**: Used as `TEAMS_APP_PASSWORD`

### Step 3: Configure Messaging Endpoint

1. In Bot settings, find **Messaging endpoint**
2. Set it to: `https://your-domain.com/webhook/teams`
3. Click **Save**

### Step 4: Enable Teams Channel

1. Go to **Channels** in the left menu
2. Click **Edit** next to Microsoft Teams
3. Ensure **Teams** is enabled (should be by default)
4. Agree to terms and click **Done**

### Step 5: Create Teams App Manifest

In the Azure Bot Service:

1. Go to **App Service Editor** (or use your preferred editor)
2. Create/update `manifest.json`:

```json
{
  "id": "{YOUR_APP_ID}",
  "version": "1.0.0",
  "name": "AIExpense",
  "description": "Track expenses through Teams",
  "icons": {
    "color": "#FF6F00",
    "outline": "#FF6F00"
  },
  "accentColor": "#FF6F00",
  "bots": [
    {
      "botId": "{YOUR_APP_ID}",
      "scopes": ["personal", "team"],
      "isNotificationOnly": false,
      "supportsFiles": false
    }
  ],
  "permissions": ["identity", "messageTeamMembers"],
  "validDomains": ["your-domain.com"]
}
```

### Step 6: Configure Environment Variables

Add these to your `.env` file:

```bash
TEAMS_APP_ID=your-app-id-here
TEAMS_APP_PASSWORD=your-app-password-here
```

### Step 7: Start the Server

```bash
export TEAMS_APP_ID=your-app-id-here
export TEAMS_APP_PASSWORD=your-app-password-here
go run ./cmd/server
```

### Step 8: Install Bot in Teams

1. Go to [App Studio in Teams](https://teams.microsoft.com/l/app/00000000-0000-0000-0000-000000000000)
   - Or search for "App Studio" in Teams
2. Click **Manifest editor**
3. Click **Import an existing app**
4. Upload your manifest file
5. Finish setup and click **Install**
6. Choose team/chat to install in
7. Click **Install**

## Architecture

The Teams integration follows the same adapter pattern as other messengers (LINE, Telegram, Discord, WhatsApp, Slack):

```
Teams User Message
        ↓
Webhook Handler (verify signature)
        ↓
Auto-signup (if new user)
        ↓
Parse Conversation (extract expenses)
        ↓
Create Expense (with AI categorization)
        ↓
Send Response (via Teams API)
```

### File Structure

```
internal/adapter/messenger/teams/
├── client.go      # Teams Bot API HTTP client
├── handler.go     # Webhook request handler
└── usecase.go     # Business logic orchestration
```

## Features

### 1. Direct Messages (1:1 Chat)

Send expense information directly to the bot:

```
User: breakfast $8 lunch $12
Bot: ✅ breakfast - $8.00 (Food)
     ✅ lunch - $12.00 (Food)
     Recorded 2 expense(s)
```

### 2. Channel Mentions

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

Expenses are automatically categorized:
- Food & Dining
- Transportation
- Shopping
- Entertainment
- Utilities
- Other

### 5. Conversation Context

The bot maintains conversation context and can process multiple messages in sequence.

### 6. Rich Formatting

Responses support Teams markdown formatting:
- **Bold**: `*text*`
- *Italic*: `_text_`
- Lists: `• item`
- Code: `` `code` ``

## Configuration

### Environment Variables

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `TEAMS_APP_ID` | Yes | Microsoft App ID | `123e4567-e89b-12d3-a456-...` |
| `TEAMS_APP_PASSWORD` | Yes | App password | `abc123~def456...` |

### Optional Configuration

To support multiple Teams workspaces, run multiple instances with different credentials.

## Webhook Setup

### Request Verification

The handler automatically verifies all Teams requests using HMAC-SHA256:

1. **Signature Check**: Validates `Authorization` header with HMAC
2. **Body Verification**: Computes signature of request body
3. **Comparison**: Ensures request came from Microsoft Teams

### Activity Types

The handler processes these activity types:

#### Message Activities
- **Type**: `message`
- **Scope**: Direct messages and mentions
- **Processing**: Automatic expense parsing and creation

#### Mention Activities
- **Condition**: Bot is mentioned (`@BotName`)
- **Processing**: Extracts message text and processes for expenses

#### Conversation Update Activities
- **Type**: `conversationUpdate`
- **Scope**: Bot added to team or chat
- **Processing**: Logs activity (optional welcome message)

#### Event Activities
- **Type**: `event`
- **Processing**: Logs for debugging

## API Reference

### Client Methods

#### SendMessage

```go
func (c *Client) SendMessage(conversationID, text string) error
```

Sends a message to a Teams conversation.

**Parameters:**
- `conversationID`: Teams conversation ID
- `text`: Message text (supports markdown)

**Example:**
```go
client.SendMessage("conversation_id", "✅ Lunch recorded - $12.00")
```

#### GetBotInfo

```go
func (c *Client) GetBotInfo() map[string]interface{}
```

Retrieves bot information.

**Returns:**
```go
map[string]interface{}{
  "app_id": "123e4567-e89b-12d3-a456-...",
  "app_name": "AIExpense Bot",
  "platform": "Microsoft Teams",
}
```

#### UpdateActivity

```go
func (c *Client) UpdateActivity(conversationID, activityID, text string) error
```

Updates an existing message in Teams.

**Parameters:**
- `conversationID`: Teams conversation ID
- `activityID`: ID of the activity to update
- `text`: New message text

#### SetServiceURL

```go
func (c *Client) SetServiceURL(serviceURL string)
```

Sets the service URL for API calls (automatically set per activity).

### UseCase Methods

#### ProcessMessage

```go
func (u *UseCase) ProcessMessage(ctx context.Context, userID, text string) error
```

Processes an incoming message for expense parsing.

**Flow:**
1. Auto-signup user (if new)
2. Parse text for expenses
3. Create expense records
4. Send confirmation

#### ProcessMention

```go
func (u *UseCase) ProcessMention(ctx context.Context, userID, text string) error
```

Processes a message where the bot is mentioned.

**Flow:**
1. Remove bot mention from text
2. Call ProcessMessage with cleaned text

## Message Flow

### Complete Flow Diagram

```
┌──────────────────┐
│  Teams User      │
│  sends message   │
└────────┬─────────┘
         │
         v
┌──────────────────────────────┐
│  POST /webhook/teams         │
│  (Teams sends activity)      │
└────────┬─────────────────────┘
         │
         v
┌──────────────────────────────┐
│  Verify Signature            │
│  (HMAC-SHA256)               │
└────────┬─────────────────────┘
         │
         v
┌──────────────────────────────┐
│  Parse Activity              │
│  (Message, Mention, etc)     │
└────────┬─────────────────────┘
         │
         v
┌──────────────────────────────┐
│  Set Service URL             │
│  (for API responses)         │
└────────┬─────────────────────┘
         │
         v
┌──────────────────────────────┐
│  Auto-Signup                 │
│  (if new user)               │
└────────┬─────────────────────┘
         │
         v
┌──────────────────────────────┐
│  Parse Conversation          │
│  (Extract expenses)          │
└────────┬─────────────────────┘
         │
         v
┌──────────────────────────────┐
│  Create Expenses             │
│  (Save to database)          │
└────────┬─────────────────────┘
         │
         v
┌──────────────────────────────┐
│  Send Response               │
│  (via Teams Bot API)         │
└────────┬─────────────────────┘
         │
         v
┌──────────────────┐
│  Teams User      │
│  sees response   │
└──────────────────┘
```

### Example Conversation

**User Message (Direct Chat):**
```
breakfast $8 lunch $12
```

**Internal Processing:**
```
1. Verify webhook signature ✓
2. Parse activity (type: message, user: U12345678)
3. Set service URL for responses
4. Auto-signup: teams_U12345678 ✓
5. Parse conversation:
   - breakfast: $8
   - lunch: $12
6. Create expenses:
   - breakfast → Food category → ID: exp_001
   - lunch → Food category → ID: exp_002
7. Send message via Teams Bot API
```

**Bot Response:**
```
✅ breakfast - $8.00 (Food)
✅ lunch - $12.00 (Food)
Recorded 2 expense(s)
```

## Troubleshooting

### Webhook Not Receiving Events

**Problem**: Handler doesn't receive any messages

**Solutions:**
1. Verify messaging endpoint is set to `https://your-domain.com/webhook/teams`
2. Check that Teams channel is enabled for the bot
3. Ensure bot is installed in your Teams workspace
4. Verify Teams App ID matches configuration
5. Check server logs for webhook delivery errors

### Signature Verification Failed

**Error**: "signature verification failed"

**Solutions:**
1. Verify `TEAMS_APP_PASSWORD` environment variable matches Azure credentials
2. Ensure app password is current (old passwords are invalidated when new ones are created)
3. Check server clock is synchronized (Azure uses timestamp validation)
4. Verify `Authorization` header is being received

### Bot Not Responding to Messages

**Problem**: Bot receives events but sends no responses

**Possible Causes:**
1. `TEAMS_APP_ID` is invalid or mismatched
2. `TEAMS_APP_PASSWORD` is expired or incorrect
3. Service URL not being set properly in activity
4. Parsing errors (check server logs)
5. Database errors storing expenses

**Debug:**
```bash
# Check bot credentials
curl -X POST https://login.microsoftonline.com/botframework.com/oauth2/v2.0/token \
  -d "grant_type=client_credentials" \
  -d "client_id=$TEAMS_APP_ID" \
  -d "client_secret=$TEAMS_APP_PASSWORD" \
  -d "scope=https://api.botframework.com/.default"
```

### Messages Cut Off or Malformed

**Solutions:**
1. Teams messages have 2000 character limit per activity
2. Use markdown for formatting: `*bold*`, `_italic_`
3. Split long messages into multiple activities
4. Check activity type is "message"

### Database Errors

**Problem**: Expenses not saving

**Solutions:**
1. Verify SQLite database is writable
2. Check `DATABASE_PATH` environment variable
3. Ensure migrations ran at startup
4. Check disk space

### App Not Installing in Teams

**Problem**: "This app couldn't be installed" error

**Solutions:**
1. Verify manifest.json is valid JSON
2. Check bot ID in manifest matches `TEAMS_APP_ID`
3. Verify valid domains include your server domain
4. Ensure Teams admin hasn't restricted bot installations
5. Try uploading to Teams App Studio again

## Best Practices

### 1. Error Handling

The bot provides clear error messages:
- If parsing fails: "I didn't find any expenses..."
- If database error: "處理失敗，請稍後重試"
- If partial failure: "X expense(s) failed to record"

### 2. Message Formatting

Use Teams markdown for readability:

```
✅ *Expense Recorded*
• Coffee - $5.00 (Food)
• Lunch - $12.00 (Food)

Total: $17.00
```

### 3. Conversation Context

Teams provides rich context:
- User information
- Conversation type (personal, group, channel)
- Channel ID and team ID
- Timestamp and thread context

### 4. Security

- Always use HTTPS for webhook
- Rotate app passwords regularly
- Use environment variables for credentials
- Never commit credentials to version control
- Validate signatures on all requests

### 5. User Experience

- Respond within 3 seconds for better UX
- Use threaded replies in channels
- Provide helpful suggestions for invalid input
- Include totals in multi-expense responses

## Advanced Usage

### Direct User ID Format

User IDs from Teams are prefixed with `teams_` in the system:

```
Teams User ID: 29:1234567890
AIExpense User ID: teams_29:1234567890
```

This allows separation from other messenger users.

### Conversation Types

Teams supports different conversation types:

- **Personal** (1:1): Direct bot chat
- **Group** (Group chat): Multiple users
- **Channel** (Team channel): Team conversations

The bot processes messages from all types when mentioned in channels.

### Rich Activity Types

Teams activities contain rich metadata:

```go
Activity {
  Type: "message",
  ChannelID: "channel_id",
  ChannelData: {
    TeamID: "team_id",
    ChannelName: "general"
  },
  ServiceURL: "https://smba.trafficmanager.net/...",
  Conversation: {
    ID: "conversation_id",
    IsGroup: false
  }
}
```

## Performance Considerations

- **Async Processing**: Webhook responses sent asynchronously
- **Timeout**: Teams expects response within 15 seconds
- **Rate Limiting**: Respect Teams API rate limits
- **Webhook Retries**: Teams retries failed webhooks
- **Concurrent Users**: Stateless handler supports horizontal scaling

## Support & Debugging

### Enable Debug Logging

Add to environment:
```bash
LOG_LEVEL=debug
```

### Teams Activity Inspector

Use [Bot Framework Inspector](https://docs.microsoft.com/en-us/azure/bot-service/bot-service-debug-bot?view=azure-bot-service-4.0):

1. Download ngrok for local testing
2. Create tunnel: `ngrok http 8080`
3. Set messaging endpoint to ngrok URL
4. Use inspector to view activity payload

### Test with Bot Emulator

- Download [Bot Framework Emulator](https://github.com/Microsoft/BotFramework-Emulator)
- Connect to: `http://localhost:8080/webhook/teams`
- Input your App ID and Password
- Test message scenarios

## Limitations & Future Enhancements

### Current Limitations

- Thread replies: Basic support
- File uploads: Not yet supported
- Card attachments: Plain text only
- Scheduled messages: Not supported
- Proactive messaging: Requires additional setup

### Planned Features

- [ ] Adaptive Cards for rich UI
- [ ] File upload support
- [ ] Scheduled expense reminders
- [ ] Budget alerts in Teams
- [ ] Team expense sharing
- [ ] Monthly reports in channels
- [ ] Integration with Teams Calendar
- [ ] Slash commands for commands
- [ ] Workflow automations

## Contributing

Found a bug or have suggestions? Open an issue in the repository.

---

**Last Updated**: January 2026
**Version**: 1.0.0
**Status**: Production Ready ✅
**Authentication**: HMAC-SHA256 Signature Verification
**API Version**: Teams Bot Framework v3
