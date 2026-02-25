# User Journey Diagrams

This document captures the primary user journeys currently supported in AIExpense.

## 1) Product-Level Journey Map

```mermaid
flowchart TD
    A[User enters via chat or web] --> B{New user?}
    B -->|Yes| C[Auto-signup and create default categories]
    B -->|No| D[Load user profile and preferences]
    C --> E[Submit expense message]
    D --> E

    E --> F[Parse message: amount, currency, category, account, date]
    F --> G[Create one or more expenses]
    G --> H{Next user goal}

    H -->|Track more| E
    H -->|View report| I[Generate report link]
    H -->|Manage data| J[Dashboard/API operations]

    I --> K[Open short link]
    K --> L[Token-authenticated dashboard session]
    L --> J

    J --> J1[Search/filter expenses]
    J --> J2[Edit/delete expenses]
    J --> J3[Manage categories]
    J --> J4[Recurring and budgets]
    J --> J5[Export/archive/restore]
```

## 2) Expense Capture Journey (Chat to Stored Expense)

```mermaid
sequenceDiagram
    participant U as User
    participant T as Chat Endpoint
    participant P as ProcessMessage
    participant X as ParseConversation
    participant C as CreateExpense
    participant R as Repository

    U->>T: "Lunch 120 TWD on credit card"
    T->>P: UserMessage
    P->>X: Parse text
    X-->>P: ParsedExpense(s)<br/>description, amount, currency, category, account, date

    loop for each parsed expense
        P->>C: CreateRequest (includes parsed fields)
        C->>R: Persist expense
        R-->>C: Created expense
        C-->>P: CreateResponse
    end

    P-->>T: Confirmation message
    T-->>U: Recorded expense result
```

## 3) Report and Dashboard Access Journey

```mermaid
sequenceDiagram
    participant U as User
    participant T as Chat Endpoint
    participant P as ProcessMessage
    participant L as Report Link Service
    participant S as Short Link Handler
    participant D as Dashboard API

    U->>T: "Show report"
    T->>P: UserMessage
    P->>L: Generate report link
    L-->>P: Short link URL
    P-->>U: Report link

    U->>S: Open short link
    S-->>U: Redirect with report token
    U->>D: API call with token
    D-->>U: Expenses and aggregates
    U->>D: Edit expense
    D-->>U: Updated expense
```

## 4) Ongoing Data Management Journey

```mermaid
flowchart LR
    A[Active user] --> B[Manage categories]
    A --> C[Track recurring expenses]
    A --> D[Set budgets and compare]
    A --> E[Search and filter history]
    A --> F[Notifications and preferences]
    A --> G[Export or archive data]

    B --> B1[Create/rename/delete/merge]
    C --> C1[Create/update/process recurring entries]
    D --> D1[Budget status and variance]
    E --> E1[Date/category/account filters]
    F --> F1[Read/mark/update notification rules]
    G --> G1[Export summary or restore archive]
```

## 5) Multi-Channel and Admin Operations

```mermaid
flowchart TD
    U1[End users] --> M{Message channel}
    M --> M1[LINE]
    M --> M2[Slack]
    M --> M3[Telegram]
    M --> M4[Discord]
    M --> M5[Teams]
    M --> M6[WhatsApp]
    M --> M7[Terminal]

    M1 --> P[Shared processing use cases]
    M2 --> P
    M3 --> P
    M4 --> P
    M5 --> P
    M6 --> P
    M7 --> P

    A1[Admin] --> A2[Admin auth]
    A2 --> A3[Analytics]
    A2 --> A4[AI cost metrics]
    A2 --> A5[Pricing sync]
```
