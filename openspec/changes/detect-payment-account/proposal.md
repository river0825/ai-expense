# Proposal: Detect Payment Account from Expense Description

## Summary
Add support for detecting the payment account (e.g., "Cash", "Credit Card", "Bank Transfer") directly from the natural language expense description using the LLM.

## Background
Currently, expenses are captured with just a description, amount, and category. Users often specify the payment method in the text (e.g., "Lunch 200 Cash"), but this information is not structured. Capturing this allows for better financial tracking.

## Goals
1. Use LLM to detect payment account/method from text.
2. Default to "Cash" if not specified.
3. Persist this information in the backend (PostgreSQL).
4. Display this information in the frontend with a premium, mobile-first UI.

## Design Strategy
- **Design System**: Generated via `ui-ux-pro-max` (Dark Mode, IBM Plex Sans, Amber/Slate palette).
- **Mobile First**: The UI will be optimized for mobile devices first, then adapted for desktop.
- **Principles**: Clean Architecture, TDD, BDD.

## Non-Goals
- Automatic reconciliation with bank APIs (this is just text tagging).
- SQLite support (deprecated).
