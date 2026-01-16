# Discord Bot Integration Guide

This guide explains how to set up and configure the AIExpense Discord Bot for tracking expenses through Discord.

## Overview

The Discord bot allows users to track their expenses directly through Discord by sending messages like:
- "breakfast $20"
- "uber $15 lunch $30"
- Multiple expenses in one message are automatically parsed

The bot integrates with AIExpense using the same expense parsing engine as LINE and Telegram, supporting natural language input, automatic categorization, and cross-platform user management.

## Prerequisites

- A Discord server (guild) where you have administrative permissions
- Discord Developer Portal access
- The AIExpense backend running with Discord support enabled
- Internet access to publicly expose your webhook endpoint

## Step 1: Create a Discord Application

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Click "New Application" button
3. Give your application a name (e.g., "AIExpense Bot")
4. Accept the terms and create the application
5. Go to the "Bot" section in the left sidebar
6. Click "Add Bot"
7. Under the TOKEN section, click "Copy" to copy your bot token

**Save this token securely** - you'll need it for configuration.

## Step 2: Configure Bot Permissions

1. In the Bot settings, scroll down to "TOKEN PERMISSIONS"
2. Select the following scopes:
   - `applications.commands`
   - `bot` (for the scope)

3. Select the following permissions:
   - `Send Messages`
   - `Read Messages/View Channels`
   - `Read Message History`

4. Copy the generated OAuth2 URL from the scopes section
5. Open it in your browser to add the bot to your Discord server

## Step 3: Enable Interaction Endpoint

1. In the Discord Developer Portal, go to your application
2. Click on "General Information"
3. Copy your "Application ID"
4. Go to the "Interactions Endpoint URL" field
5. Enter your webhook URL: `https://your-domain.com/webhook/discord`
6. Discord will send a POST request to verify the endpoint
7. Save the changes

**Note**: Your endpoint must be publicly accessible and respond to Discord's verification ping.

## Step 4: Configure Environment Variables

Add the following to your `.env` file or set as environment variables:

```bash
# Discord Bot Token (from Step 1)
DISCORD_BOT_TOKEN=your_bot_token_here

# Server configuration
SERVER_PORT=8080

# Database
DATABASE_PATH=./aiexpense.db

# AI Provider
GEMINI_API_KEY=your_gemini_key
AI_PROVIDER=gemini

# LINE (still required for main app)
LINE_CHANNEL_TOKEN=your_line_token
LINE_CHANNEL_ID=your_line_channel_id

# Admin API Key for metrics
ADMIN_API_KEY=your_admin_key
```

## Step 5: Deploy and Test

1. Start the AIExpense server:
```bash
go run ./cmd/server/main.go
```

2. You should see in the logs:
```
Discord webhook enabled at /webhook/discord
```

3. In your Discord server, type a message to test:
```
breakfast $20
```

4. The bot should respond with a confirmation message

## How It Works

### User Identification

- Users are automatically registered on their first message
- User ID format: `discord_{user_id}`
- Each user gets a set of default expense categories
- Categories can be customized per user

### Message Processing Flow

1. **Receive**: Bot receives user message via Discord interaction webhook
2. **Parse**: Message is sent to AI service (Gemini) for expense extraction
   - Supports natural language: "早餐$20午餐$30"
   - Supports English: "breakfast $20 lunch $30"
3. **Categorize**: AI suggests categories for each expense
4. **Create**: Expenses are saved to the database
5. **Respond**: Bot sends confirmation message back to user

### Supported Message Formats

```
Single expense:
- "breakfast $20"
- "coffee 5"
- "dinner $45.99"

Multiple expenses:
- "breakfast $20 lunch $30 dinner $50"
- "coffee 5 uber 15 groceries 100"

Natural language (with Gemini API):
- "早餐$20午餐$30" (Chinese: breakfast $20 lunch $30)
- "朝食20円昼食30円" (Japanese)
```

### Error Handling

If parsing fails, the bot will respond with:
- "No valid expense items found. Please provide an amount and item"
- If individual expenses fail to save: "{item} (save failed)"

## Configuration Options

### Optional: Verify Bot Connection

```bash
curl -X POST http://localhost:8080/webhook/discord \
  -H "Content-Type: application/json" \
  -d '{
    "type": 1,
    "id": "test",
    "token": "test"
  }'
```

The bot should respond with a PONG (interaction type 1).

## Testing the Bot

### Test Expense Entry

1. Send a message in Discord:
```
coffee $5
```

2. Expected response:
```
Expense saved: coffee ($5.00) - 2024-01-16
```

### Test Multiple Expenses

```
breakfast $10 lunch $15 coffee $5
```

Expected response:
```
Expense saved: breakfast ($10.00) - 2024-01-16
Expense saved: lunch ($15.00) - 2024-01-16
Expense saved: coffee ($5.00) - 2024-01-16
```

### Check Your Expenses

Use the REST API to query your expenses:
```bash
curl -X GET http://localhost:8080/api/expenses \
  -H "X-API-Key: your_admin_key"
```

## Troubleshooting

### Bot Doesn't Respond

1. **Check webhook URL**: Ensure your endpoint URL is correct and publicly accessible
   ```bash
   curl -X POST https://your-domain/webhook/discord
   ```

2. **Check bot token**: Verify DISCORD_BOT_TOKEN is set correctly
   ```bash
   echo $DISCORD_BOT_TOKEN
   ```

3. **Check server logs**: Look for error messages in application logs
   ```bash
   grep -i discord /var/log/aiexpense.log
   ```

4. **Verify permissions**: Ensure bot has "Send Messages" permission in the channel

### Interaction Endpoint URL Not Responding

1. Ensure your server is running and accessible from the internet
2. Verify firewall rules allow inbound traffic on your port
3. Check that the URL is correctly formatted (https://, no trailing slash)
4. Test the endpoint manually:
   ```bash
   curl -v -X POST https://your-domain/webhook/discord
   ```

### Permission Errors

If you see "Missing Permission" errors:
1. Go to Discord Developer Portal
2. In your application, go to OAuth2 → URL Generator
3. Add the bot to your server with these permissions:
   - View Channels
   - Send Messages
   - Read Message History

### AI Service Errors

If message parsing fails:
1. Check GEMINI_API_KEY is set and valid
2. Verify AI_PROVIDER is set to "gemini" (or your configured provider)
3. Check API quotas and rate limits
4. Review error logs for detailed error messages

## Comparison with Other Messengers

| Feature | Discord | LINE | Telegram |
|---------|---------|------|----------|
| **Setup Complexity** | Medium | Low | Low |
| **User Base** | Gaming/Dev Communities | Asia-focused | Global |
| **Message Format** | Interactions | Reply Messages | Messages |
| **Webhook Type** | Interaction Endpoint | Webhook | Polling/Webhook |
| **Authentication** | Bot Token | Channel Token | Bot Token |
| **Cost** | Free | Free | Free |
| **Latency** | Low | Low | Low |

## Advanced Configuration

### Multiple Bots

To run multiple Discord bots:
1. Create separate Discord applications in Developer Portal
2. For each bot, configure a unique webhook path:
   ```bash
   # In a modified handler
   mux.HandleFunc("POST /webhook/discord/bot1", handler1.HandleWebhook)
   mux.HandleFunc("POST /webhook/discord/bot2", handler2.HandleWebhook)
   ```

### Rate Limiting

Discord enforces rate limits:
- Normal: 5 requests per 5 seconds
- Slash commands: Limited by Discord globally

The AIExpense bot respects these limits automatically through Go's http.Client.

### Custom Message Formatting

To customize bot responses, modify:
- `internal/adapter/messenger/discord/usecase.go` - Message composition
- `internal/usecase/create_expense.go` - Confirmation message format

## Security Considerations

1. **Protect your bot token**: Never commit it to version control
   - Use environment variables
   - Use secrets management (AWS Secrets Manager, Vault, etc.)

2. **Webhook validation**: Discord validates webhook requests with:
   - Public key verification (built into the handler)
   - Request signature validation

3. **Data privacy**:
   - User IDs are stored with "discord_" prefix
   - Messages are not logged or stored
   - Only extracted expenses are persisted

## Support and Resources

- [Discord Developer Documentation](https://discord.com/developers/docs)
- [Discord.py Documentation](https://discordpy.readthedocs.io/)
- [AIExpense Documentation](./README.md)

## Additional Features Coming Soon

- [ ] Slash commands for expense entry (more structured format)
- [ ] Button-based category selection
- [ ] Expense editing via reactions
- [ ] Monthly reports as Discord embeds
- [ ] Multi-server configuration
- [ ] Rich message formatting with embeds

---

**Last Updated**: 2024-01-16
**Version**: 1.0.0
