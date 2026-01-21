# WhatsApp Business API Integration Guide

This guide explains how to set up and configure the AIExpense WhatsApp Business Bot for tracking expenses through WhatsApp.

## Overview

The WhatsApp bot allows users to track their expenses directly through WhatsApp by sending messages like:
- "breakfast $20"
- "coffee 5 uber 15"
- "晚餐$100" (in various languages)

The bot integrates with AIExpense using the same expense parsing engine as LINE, Telegram, and Discord, supporting natural language input, automatic categorization, and cross-platform user management.

## Prerequisites

- A Meta Business Account
- WhatsApp Business Account approval
- Administrator access to manage your WhatsApp Business Account
- Existing phone number to use for the WhatsApp bot
- The AIExpense backend running with WhatsApp support enabled
- Internet access to publicly expose your webhook endpoint

## Step 1: Set Up WhatsApp Business Account

### Create a Meta Business Account

1. Go to [Meta Business Suite](https://business.facebook.com/)
2. Create a new business account or use an existing one
3. Verify your business information
4. Accept the Terms of Service

### Get WhatsApp Business Account Approval

1. In Meta Business Suite, navigate to "Apps"
2. Create a new app or use an existing one
3. Add WhatsApp product to your app
4. Follow Meta's approval process:
   - Submit business information
   - Provide use case details
   - Explain how you'll use the API
   - Wait for Meta's review (typically 1-7 days)

## Step 2: Create WhatsApp Business Phone Number

1. In Meta Business Suite, go to "Accounts" → "WhatsApp Accounts"
2. Create a new WhatsApp Business Account
3. Add a phone number to use as your bot
4. Complete phone verification:
   - Meta will send a verification code via SMS
   - Enter the code to verify
   - If you don't receive the code, you can receive it via voice call

**Note**: You can use:
- Your existing business phone number
- A dedicated phone number for the bot
- A virtual number from services like Twilio, Vonage, etc.

## Step 3: Generate API Credentials

1. In Meta Business Suite, go to "Settings" → "Apps and Websites"
2. Open your app settings
3. Go to "WhatsApp" → "API Setup"
4. You'll find:
   - **Phone Number ID**: A unique identifier for your WhatsApp Business Phone Number
   - **Business Account ID**: Your WhatsApp Business Account ID

5. Generate an access token:
   - Go to "Settings" → "User Access Tokens"
   - Create a new token with permissions: `whatsapp_business_messaging`
   - Copy the token (keep it secure!)

6. Find your app secret (webhook verification):
   - Go to "Settings" → "Basic"
   - Copy your "App Secret"

**Save these securely**:
- Phone Number ID
- Access Token
- App Secret

## Step 4: Configure Environment Variables

Add the following to your `.env` file or set as environment variables:

```bash
# WhatsApp Business API credentials
WHATSAPP_PHONE_NUMBER_ID=your_phone_number_id
WHATSAPP_ACCESS_TOKEN=your_access_token
WHATSAPP_APP_SECRET=your_app_secret

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

## Step 5: Configure Webhook

1. In Meta App Dashboard, go to "Settings" → "Basic"
2. Under "Products", find "WhatsApp"
3. Click "Configure" under "Webhooks"
4. Set your **Callback URL**: `https://your-domain.com/webhook/whatsapp`
5. Set **Verify Token**: A random string you'll use to verify requests (e.g., "verify_token")
6. Select **Webhook Fields**:
   - ✅ messages
   - ✅ message_status
   - ✅ message_template_status_update

7. Click "Verify and Save"
   - Meta will send a verification request to your endpoint
   - Your bot should respond with the challenge code

## Step 6: Deploy and Test

1. Start the AIExpense server:
```bash
go run ./cmd/server/main.go
```

2. You should see in the logs:
```
WhatsApp webhook enabled at /webhook/whatsapp
```

3. In WhatsApp, message your bot's phone number with a test message:
```
breakfast $20
```

4. The bot should respond with a confirmation message

## How It Works

### User Identification

- Users are identified by their WhatsApp phone number
- User ID format: `{phone_number}` (e.g., `12025551234`)
- Each user gets a set of default expense categories
- Categories can be customized per user

### Message Processing Flow

1. **Receive**: Bot receives user message via WhatsApp webhook
2. **Parse**: Message is sent to AI service (Gemini) for expense extraction
   - Supports natural language: "breakfast $20 lunch $30"
   - Supports multiple currencies and languages
3. **Categorize**: AI suggests categories for each expense
4. **Create**: Expenses are saved to the database
5. **Respond**: Bot sends confirmation message back via WhatsApp

### Supported Message Types

```
Text messages:
- "breakfast $20"
- "coffee 5"
- "dinner $45.99"

Multiple expenses:
- "breakfast $20 lunch $30 dinner $50"
- "coffee 5 uber 15 groceries 100"

Natural language (with Gemini API):
- "早餐$20午餐$30" (Chinese)
- "朝食20円昼食30円" (Japanese)
- "petit-déjeuner 20€ déjeuner 30€" (French)
```

## Error Handling

If parsing fails, the bot will respond with:
- "No valid expense items found. Please provide an amount and item, for example: breakfast $20"
- If individual expenses fail to save: "{item} (save failed)"

## Configuration Options

### Webhook Signature Verification

The WhatsApp handler automatically verifies webhook signatures using:
- Algorithm: HMAC-SHA256
- Secret: Your app secret
- Header: `X-Hub-Signature-256`

### Response Timeouts

WhatsApp expects webhook responses within 30 seconds. The current implementation:
- Acknowledges requests immediately (HTTP 200)
- Processes messages asynchronously
- Ensures all responses happen within timeout

## Testing the Bot

### Test Expense Entry

1. Send a WhatsApp message to your bot:
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

### Webhook Verification Fails

1. **Check callback URL**: Ensure your URL is correct and publicly accessible
   ```bash
   curl -X GET 'https://your-domain/webhook/whatsapp?hub.mode=subscribe&hub.challenge=test&hub.verify_token=verify_token'
   ```

2. **Check app secret**: Verify WHATSAPP_APP_SECRET is set correctly
   ```bash
   echo $WHATSAPP_APP_SECRET
   ```

3. **Verify token mismatch**: Ensure the token you configured matches your code
   - Default in handler: `"verify_token"`
   - Change the value in `handler.go` line 138 if needed

### Bot Doesn't Respond

1. **Check credentials**:
   - Phone Number ID is correct
   - Access Token is valid and not expired
   - Both are from the same WhatsApp Business Account

2. **Check permissions**:
   - Access token has `whatsapp_business_messaging` permission
   - Business account has message sending enabled

3. **Check server logs**:
   ```bash
   grep -i whatsapp /var/log/aiexpense.log
   ```

4. **Test message sending manually**:
   ```bash
   curl -X POST https://graph.instagram.com/v18.0/{PHONE_NUMBER_ID}/messages \
     -H "Authorization: Bearer {ACCESS_TOKEN}" \
     -H "Content-Type: application/json" \
     -d '{
       "messaging_product": "whatsapp",
       "to": "1234567890",
       "type": "text",
       "text": {
         "body": "Test message"
       }
     }'
   ```

### Signature Verification Failed

If you see "Webhook signature verification failed":

1. Verify webhook secret is correct in Meta Dashboard
2. Check that the X-Hub-Signature-256 header is being sent
3. Ensure the signature algorithm matches (HMAC-SHA256)
4. Check request body hasn't been modified

### Message Not Delivered

1. **Check phone number format**:
   - Should be valid, international format
   - Remove any hyphens or spaces
   - Example: `12025551234` (not `+1 (202) 555-1234`)

2. **Check quality rating**:
   - Go to Meta Business Suite
   - Check your WhatsApp account's "Quality Rating"
   - Low quality rating can delay/block messages
   - Improve by reducing spam reports and opt-outs

3. **Check rate limits**:
   - Meta enforces rate limits on business messages
   - Default: ~60 messages per second per business account
   - Conversation state is tracked for cheaper messaging

### AI Service Errors

If message parsing fails:
1. Check GEMINI_API_KEY is set and valid
2. Verify AI_PROVIDER is set to "gemini"
3. Check API quotas and rate limits
4. Review error logs for detailed error messages

## Comparison with Other Messengers

| Feature | WhatsApp | Telegram | Discord | LINE |
|---------|----------|----------|---------|------|
| **Setup Complexity** | High | Low | Medium | Low |
| **User Base** | 2B+ (global) | 500M+ (global) | 150M+ (gaming) | 85M+ (Asia) |
| **Message Format** | Text/Media | Messages | Interactions | Messages |
| **Webhook Type** | Webhook | Polling | Interaction | Webhook |
| **Authentication** | Access Token | Bot Token | Bot Token | Channel Token |
| **Cost** | Paid (Meta API) | Free | Free | Free |
| **Latency** | ~5s avg | <1s | <1s | <1s |
| **Phone Verification** | Required | Not required | Not required | Not required |

## Advanced Configuration

### Media Support (Future)

The client includes placeholder for media upload:
```go
// To implement in future:
func (c *Client) UploadMedia(ctx context.Context, mediaURL, mediaType string) (string, error) {
    // Implementation for image/document/audio support
}
```

### Rich Message Formatting

WhatsApp supports:
- **Text messages** (currently implemented)
- **Media messages** (images, documents, audio, video)
- **Template messages** (pre-approved message templates)
- **Interactive messages** (buttons, lists)

To implement these:
1. Modify `SendMessageRequest` struct
2. Add new message type handlers
3. Update webhook payload parsing

### Multi-Number Setup

To run multiple WhatsApp bots:
1. Create separate WhatsApp Business Phone Numbers
2. For each number, get Phone Number ID and Access Token
3. Modify environment:
   ```bash
   WHATSAPP_PHONE_NUMBER_IDS=id1,id2,id3
   WHATSAPP_ACCESS_TOKENS=token1,token2,token3
   ```
4. Update server initialization to create multiple handlers

## Security Considerations

1. **Protect your tokens**:
   - Never commit to version control
   - Use environment variables or secrets management
   - Rotate tokens periodically

2. **Webhook signature validation**:
   - Always verify X-Hub-Signature-256 header
   - Use your app secret for verification
   - Reject requests with invalid signatures

3. **Data privacy**:
   - Phone numbers are stored as user identifiers
   - Messages are not logged or stored
   - Only extracted expenses are persisted
   - Comply with WhatsApp privacy policies

4. **Rate limiting**:
   - WhatsApp enforces rate limits per business account
   - Monitor your Quality Rating
   - Handle rate limit responses gracefully

## Meta API Documentation

- [WhatsApp Business API Docs](https://developers.facebook.com/docs/whatsapp/cloud-api/get-started)
- [Message Types Reference](https://developers.facebook.com/docs/whatsapp/cloud-api/messages)
- [Webhook Reference](https://developers.facebook.com/docs/whatsapp/webhooks/components)
- [Error Codes](https://developers.facebook.com/docs/whatsapp/api/errors)

## Additional Features Coming Soon

- [ ] Media message support (images, documents, video)
- [ ] Template messages (pre-approved message sets)
- [ ] Interactive messages (buttons, lists)
- [ ] Message reactions
- [ ] Group message support
- [ ] Expense report delivery via WhatsApp
- [ ] Expense editing through WhatsApp replies

---

**Last Updated**: 2024-01-16
**Version**: 1.0.0
**Status**: Production Ready
