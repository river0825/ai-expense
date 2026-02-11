## MODIFIED Requirements

### Requirement: Send Clarification Requests
The system SHALL ask users for missing information before saving incomplete expenses or applying incomplete settings actions, and SHALL persist pending conversation state to interpret the next user reply.

#### Scenario: Missing description
- **WHEN** user sends "$100" without description
- **THEN** system responds: "這$100是什麼消費?" and waits for user response

#### Scenario: Missing amount
- **WHEN** user sends "買菜" without amount
- **THEN** system responds: "買菜花了多少錢?" and waits for user response

#### Scenario: Ambiguous category
- **WHEN** system suggests category but user doesn't confirm
- **THEN** system asks: "我建議是 [分類名] 分類，可以嗎?" with Yes/No options

#### Scenario: Missing currency for setting change
- **WHEN** user requests currency change without a resolvable target currency
- **THEN** system asks: "你要切換成哪個幣別？"
- **AND** system stores pending state so the next user message is interpreted as currency answer
