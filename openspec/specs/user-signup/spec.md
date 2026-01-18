# user-signup Specification

## Purpose
TBD - created by archiving change add-conversational-expense-tracker. Update Purpose after archive.
## Requirements
### Requirement: Auto-Signup on First Message
The system SHALL automatically create a user account when a user sends their first message to the bot via any supported messenger (LINE, Telegram, etc.).

#### Scenario: LINE user sends first message
- **WHEN** a LINE user sends any message to the bot for the first time
- **THEN** system creates user record with user_id (from LINE), messenger_type='line'
- **AND** system initializes default categories (Food, Transport, Shopping, Entertainment, Other)
- **AND** system responds to the message (auto-signup is transparent to user)

#### Scenario: Telegram user sends first message
- **WHEN** a Telegram user sends any message to the bot for the first time
- **THEN** system creates user record with user_id (from Telegram), messenger_type='telegram'
- **AND** system initializes default categories
- **AND** system responds to the message

#### Scenario: User already exists
- **WHEN** an existing user sends a message
- **THEN** system recognizes user and skips signup
- **AND** proceeds with normal message handling

#### Scenario: Duplicate signup attempt
- **WHEN** same user sends multiple messages in quick succession before first signup completes
- **THEN** system handles race condition gracefully (idempotent)
- **AND** exactly one user record is created

### Requirement: Initialize User State on Signup
The system SHALL initialize necessary user configuration when account is created.

#### Scenario: Default categories created
- **WHEN** new user is registered
- **THEN** system creates 5 default categories: Food, Transport, Shopping, Entertainment, Other
- **AND** all categories are marked as is_default=true

#### Scenario: User can immediately use expense tracking
- **WHEN** user signs up and sends expense message (e.g., "早餐$20")
- **THEN** system can categorize and save expense without further user setup

### Requirement: Support Multiple Messengers on Signup
The system SHALL track which messenger each user is using and support independent signup flows per messenger type.

#### Scenario: Same user on multiple messengers
- **WHEN** same user (e.g., by phone number or manual linking) uses bot on LINE and Telegram
- **THEN** system treats as two separate users with two user_ids
- **AND** data is isolated per messenger/user_id combination

#### Scenario: Messenger type tracked
- **WHEN** user is registered
- **THEN** user record includes messenger_type field ('line', 'telegram', etc.)

