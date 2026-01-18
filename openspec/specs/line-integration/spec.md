# line-integration Specification

## Purpose
TBD - created by archiving change add-conversational-expense-tracker. Update Purpose after archive.
## Requirements
### Requirement: Receive Messages from LINE Messaging API
The system SHALL receive and process incoming messages from LINE users through the Messaging API webhook.

#### Scenario: Webhook receives message
- **WHEN** user sends text message in LINE bot chat
- **THEN** system receives webhook event with user ID, message text, and timestamp

#### Scenario: Support various message types
- **WHEN** user sends text, but also possibly sends stickers or locations
- **THEN** system prioritizes text messages and gracefully ignores unsupported types

#### Scenario: Webhook responds within timeout
- **WHEN** LINE sends webhook to bot endpoint
- **THEN** system responds with HTTP 200 within 3 seconds

### Requirement: Send Confirmation Messages
The system SHALL send confirmation messages to LINE users when expenses are successfully saved.

#### Scenario: Single expense confirmation
- **WHEN** user sends "早餐$20" successfully
- **THEN** system responds: "早餐 20元，已儲存"

#### Scenario: Multiple expenses confirmation
- **WHEN** user sends "早餐$20午餐$30加油$200"
- **THEN** system responds in single message:
```
早餐 20元，已儲存
午餐 30元，已儲存
加油 200元，已儲存
```

#### Scenario: Category included in confirmation
- **WHEN** expense is saved with category (e.g., Food)
- **THEN** system responds: "早餐 20元 [食物]，已儲存"

#### Scenario: Confirmation includes date when non-today
- **WHEN** user sends "昨天買水果$300"
- **THEN** system responds: "水果 300元 (昨天)，已儲存"

### Requirement: Send Clarification Requests
The system SHALL ask users for missing information before saving incomplete expenses.

#### Scenario: Missing description
- **WHEN** user sends "$100" without description
- **THEN** system responds: "這$100是什麼消費?" and waits for user response

#### Scenario: Missing amount
- **WHEN** user sends "買菜" without amount
- **THEN** system responds: "買菜花了多少錢?" and waits for user response

#### Scenario: Ambiguous category
- **WHEN** system suggests category but user doesn't confirm
- **THEN** system asks: "我建議是 [分類名] 分類，可以嗎?" with Yes/No options

### Requirement: Send Report Messages
The system SHALL format and send expense reports as LINE messages that are easy to read.

#### Scenario: Daily summary
- **WHEN** user requests "報表" or "report"
- **THEN** system sends single message with formatted summary:
```
今天的消費報告
============
總額: $XXX

分類統計:
• 食物: $XX
• 交通: $XX
• 購物: $XX
```

#### Scenario: Report includes currency symbol
- **WHEN** system generates report
- **THEN** system uses appropriate currency notation ($ or 元 or other as configured)

#### Scenario: Message length optimization
- **WHEN** report data is extensive
- **THEN** system fits within single LINE message (2500 character limit) by prioritizing recent data

### Requirement: Handle User Errors Gracefully
The system SHALL provide helpful error messages when user input cannot be parsed or saved.

#### Scenario: Invalid input format
- **WHEN** user sends malformed input that parser cannot process
- **THEN** system responds with friendly suggestion: "看不懂呢，請給我金額和品項，例如：早餐$20"

#### Scenario: Database error
- **WHEN** system encounters database error during save
- **THEN** system responds to user: "儲存失敗，請稍後重試" (without exposing technical error details)

#### Scenario: Parse timeout
- **WHEN** parser takes too long to process input
- **THEN** system responds: "處理超時，請簡化輸入並重試"

### Requirement: LINE User Authentication
The system SHALL verify that webhook messages come from LINE and maintain user context across messages.

#### Scenario: Webhook signature verification
- **WHEN** bot receives webhook
- **THEN** system verifies X-Line-Signature header matches HMAC-SHA256 of request body

#### Scenario: Maintain user session
- **WHEN** user sends multiple messages in conversation
- **THEN** system maintains user context (recent expenses, pending actions) within same chat

#### Scenario: Multiple users
- **WHEN** different users interact with bot
- **THEN** system isolates expense data per user ID and never mixes records

