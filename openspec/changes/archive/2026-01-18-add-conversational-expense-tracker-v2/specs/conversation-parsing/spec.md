## ADDED Requirements

### Requirement: Parse Expense Descriptions from Natural Language (AI-Powered)
The system SHALL extract expense descriptions and amounts from natural language user input using Gemini 2.5 lite AI model, supporting flexible formats and natural phrasing.

#### Scenario: Single expense entry
- **WHEN** user sends "早餐$20"
- **THEN** system extracts description="早餐" and amount=20

#### Scenario: Multiple consecutive expenses
- **WHEN** user sends "早餐$20午餐$30加油$200"
- **THEN** system extracts three entries: (早餐, 20), (午餐, 30), (加油, 200)

#### Scenario: Expenses with extra text
- **WHEN** user sends "買了早餐花了$50"
- **THEN** system extracts description="早餐" and amount=50

#### Scenario: Invalid format
- **WHEN** user sends "abc def" (no amounts)
- **THEN** system returns error asking user to provide amounts with $

### Requirement: Parse Relative Dates from Natural Language
The system SHALL extract date information from natural language text, defaulting to today if no date is specified.

#### Scenario: Explicit relative date
- **WHEN** user sends "昨天買水果$300"
- **THEN** system extracts date=yesterday and expense (水果, 300)

#### Scenario: Week-relative dates
- **WHEN** user sends "上週買的衣服$500"
- **THEN** system extracts date=7 days ago and expense (衣服, 500)

#### Scenario: No date specified
- **WHEN** user sends "早餐$20" (no date mention)
- **THEN** system defaults to today's date

#### Scenario: Unsupported date format
- **WHEN** user sends "1/2/2024買東西$100" (ambiguous format)
- **THEN** system prompts user to clarify or defaults to today

### Requirement: Validate Parsed Expenses
The system SHALL validate that each parsed expense has a valid amount and description.

#### Scenario: Valid amounts
- **WHEN** system parses "早餐$20午餐$30.50"
- **THEN** system accepts both integer and decimal amounts

#### Scenario: Missing description
- **WHEN** system parses "$100" (amount only, no description)
- **THEN** system asks user "這$100是什麼消費?" to clarify

#### Scenario: Missing amount
- **WHEN** system parses "買菜" (description only, no amount)
- **THEN** system asks user "買菜花了多少錢?" to clarify

#### Scenario: Negative amounts
- **WHEN** user sends "退貨-$50" or "refund -50"
- **THEN** system interprets as negative amount (refund/credit)
