# reporting Specification

## Purpose
TBD - created by archiving change add-conversational-expense-tracker. Update Purpose after archive.
## Requirements
### Requirement: Generate Summary Report by Date Range
The system SHALL generate a concise summary of total expenses within a specified date range.

#### Scenario: Report for today
- **WHEN** user sends "report" or "今天的報告" or "報表"
- **THEN** system responds with today's total: "今天共消費: $XXX"

#### Scenario: Report for date range
- **WHEN** user sends "report this week" or "這週的報告"
- **THEN** system responds with weekly total and breakdown by category

#### Scenario: Report for month
- **WHEN** user sends "report this month" or "本月報告"
- **THEN** system responds with month total and breakdown by category

#### Scenario: Custom date range
- **WHEN** user sends "report from 2024-01-01 to 2024-01-31"
- **THEN** system responds with total for that range

#### Scenario: No expenses in range
- **WHEN** user requests report for date range with no expenses
- **THEN** system responds "此期間內沒有消費"

### Requirement: Generate Category Breakdown Report
The system SHALL show expense totals grouped by category.

#### Scenario: Category breakdown for period
- **WHEN** user sends "breakdown" or "分類統計" or "category report"
- **THEN** system responds with total by category for the period:
```
分類統計:
- 食物: $XXX (N筆)
- 交通: $XXX (N筆)
- 購物: $XXX (N筆)
- 娛樂: $XXX (N筆)
```

#### Scenario: Largest spending category
- **WHEN** user sends "which category spent most" or "花最多的分類是什麼"
- **THEN** system identifies and highlights highest spending category

#### Scenario: Category report with no expenses
- **WHEN** user requests category breakdown with no expenses in period
- **THEN** system responds "此期間內沒有消費"

### Requirement: Generate Itemized Expense List
The system SHALL list individual expenses within a time range or category.

#### Scenario: List all expenses
- **WHEN** user sends "list" or "清單" or "show all"
- **THEN** system responds with itemized list grouped by date, showing description, amount, and category

#### Scenario: List by category
- **WHEN** user sends "list food" or "顯示食物類的消費"
- **THEN** system lists all Food category expenses with totals

#### Scenario: Format for readability
- **WHEN** system generates list
- **THEN** all items fit in single LINE message with clear formatting (bullets, line breaks, totals)

### Requirement: Format Reports as Single Consolidated Message
The system SHALL deliver all report data in a single LINE message without requiring multiple API calls or user scrolling through separate responses.

#### Scenario: Report consolidation
- **WHEN** user requests complex report (summary + breakdown)
- **THEN** system responds with single message containing both summary and category breakdown

#### Scenario: Very long reports
- **WHEN** report would exceed LINE message character limit
- **THEN** system prioritizes most recent data and offers option to "view full report" with additional request

#### Scenario: Multiple expense entries
- **WHEN** user creates multiple expenses in sequence
- **THEN** system acknowledges all entries in single response message

