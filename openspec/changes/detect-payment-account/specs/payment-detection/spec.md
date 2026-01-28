# Spec: Payment Account Detection

## ADDED Requirements

### Requirement: Payment Method Extraction
The system SHALL detect the payment account or method from the expense description text using an LLM.

#### Scenario: Explicit Cash
Given the user input "Breakfast cash 200"
When the system processes the input
Then the expense payment method should be "Cash"
And the amount should be 200

#### Scenario: Explicit Credit Card
Given the user input "Gas Taishin Credit Card 1599"
When the system processes the input
Then the expense payment method should be "Taishin Credit Card"
And the amount should be 1599

#### Scenario: Implicit Default
Given the user input "Run 1599"
When the system processes the input
Then the expense payment method should be "Cash" (default)

#### Scenario: Keyword ordering variation 1
Given the user input "Taishin Credit Card Gas 1599"
When the system processes the input
Then the expense payment method should be "Taishin Credit Card"

#### Scenario: Keyword ordering variation 2
Given the user input "1599 Taishin Credit Card Gas"
When the system processes the input
Then the expense payment method should be "Taishin Credit Card"
