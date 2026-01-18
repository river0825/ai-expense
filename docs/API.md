# AIExpense API Documentation

## Overview

The AIExpense API is a RESTful service for managing expenses through natural language conversation. It supports multiple messenger platforms (LINE, Telegram, Slack, Teams, Discord, WhatsApp) and provides comprehensive expense tracking, categorization, reporting, and analytics capabilities.

**OpenAPI Specification**: `openapi.yaml` (root directory)

## Quick Start

### Base URL

- **Development**: `http://localhost:8080`
- **Production**: `https://api.aiexpense.app`

### Authentication

Most endpoints are public, but admin/metrics endpoints require an API key:

```bash
X-API-Key: your-admin-api-key
```

## Core Endpoints

### User Management

#### Auto-Signup User
**POST** `/api/users/auto-signup`

Automatically register a new user or retrieve existing user from any messenger platform.

```bash
curl -X POST http://localhost:8080/api/users/auto-signup \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "line_u123456789",
    "messenger_type": "line"
  }'
```

**Request Body**:
```json
{
  "user_id": "string (required)",
  "messenger_type": "string (required) - enum: line, telegram, slack, teams, discord, whatsapp"
}
```

**Response** (200 OK):
```json
{
  "status": "success",
  "message": "User signed up successfully"
}
```

### Expense Management

#### Parse Natural Language Expenses
**POST** `/api/expenses/parse`

Parse natural language text to extract expenses using AI.

```bash
curl -X POST http://localhost:8080/api/expenses/parse \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "line_u123456789",
    "text": "早餐$20午餐$30晚餐$50"
  }'
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": [
    {
      "description": "breakfast",
      "amount": 20,
      "date": "2024-01-18T08:00:00Z"
    },
    {
      "description": "lunch",
      "amount": 30,
      "date": "2024-01-18T12:00:00Z"
    },
    {
      "description": "dinner",
      "amount": 50,
      "date": "2024-01-18T18:00:00Z"
    }
  ]
}
```

#### Create Expense
**POST** `/api/expenses`

Create a new expense record.

```bash
curl -X POST http://localhost:8080/api/expenses \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "line_u123456789",
    "description": "breakfast at cafe",
    "amount": 20.50,
    "category_id": "cat_food",
    "expense_date": "2024-01-18T08:00:00Z"
  }'
```

**Response** (201 Created):
```json
{
  "status": "success",
  "data": {
    "id": "exp_xyz123",
    "user_id": "line_u123456789",
    "description": "breakfast at cafe",
    "amount": 20.50,
    "category_id": "cat_food",
    "created_at": "2024-01-18T08:00:00Z"
  }
}
```

#### List Expenses
**GET** `/api/expenses`

Retrieve expenses with optional filtering.

```bash
curl "http://localhost:8080/api/expenses?user_id=line_u123456789&from=2024-01-01&to=2024-01-31&category_id=cat_food"
```

**Query Parameters**:
- `user_id` (required): User ID
- `from` (optional): Start date (ISO 8601)
- `to` (optional): End date (ISO 8601)
- `category_id` (optional): Filter by category

**Response** (200 OK):
```json
{
  "status": "success",
  "data": [
    {
      "id": "exp_xyz123",
      "user_id": "line_u123456789",
      "description": "breakfast",
      "amount": 20,
      "category_id": "cat_food",
      "expense_date": "2024-01-18T08:00:00Z",
      "created_at": "2024-01-18T08:00:00Z"
    }
  ]
}
```

#### Update Expense
**PUT** `/api/expenses/{expense_id}`

```bash
curl -X PUT http://localhost:8080/api/expenses/exp_xyz123 \
  -H "Content-Type: application/json" \
  -d '{
    "description": "breakfast at nice cafe",
    "amount": 22.50
  }'
```

#### Delete Expense
**DELETE** `/api/expenses/{expense_id}`

```bash
curl -X DELETE http://localhost:8080/api/expenses/exp_xyz123
```

#### Search Expenses
**GET** `/api/expenses/search`

Search expenses by keyword.

```bash
curl "http://localhost:8080/api/expenses/search?user_id=line_u123456789&q=breakfast"
```

#### Filter Expenses
**GET** `/api/expenses/filter`

Filter expenses by multiple criteria.

```bash
curl "http://localhost:8080/api/expenses/filter?user_id=line_u123456789&min_amount=10&max_amount=50&category_id=cat_food"
```

### Category Management

#### List Categories
**GET** `/api/categories`

```bash
curl "http://localhost:8080/api/categories?user_id=line_u123456789"
```

#### Create Category
**POST** `/api/categories`

```bash
curl -X POST http://localhost:8080/api/categories \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "line_u123456789",
    "name": "Entertainment"
  }'
```

#### Update Category
**PUT** `/api/categories/{category_id}`

#### Delete Category
**DELETE** `/api/categories/{category_id}`

### Metrics & Analytics

#### Daily Active Users
**GET** `/api/metrics/dau`

Requires admin API key.

```bash
curl http://localhost:8080/api/metrics/dau \
  -H "X-API-Key: admin-key-123"
```

#### Expense Summary
**GET** `/api/metrics/expenses-summary`

```bash
curl "http://localhost:8080/api/metrics/expenses-summary?user_id=line_u123456789" \
  -H "X-API-Key: admin-key-123"
```

#### Growth Metrics
**GET** `/api/metrics/growth`

```bash
curl http://localhost:8080/api/metrics/growth \
  -H "X-API-Key: admin-key-123"
```

### Reports & Export

#### Generate Report
**POST** `/api/reports/generate`

```bash
curl -X POST http://localhost:8080/api/reports/generate \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "line_u123456789",
    "from": "2024-01-01T00:00:00Z",
    "to": "2024-01-31T23:59:59Z"
  }'
```

#### Export Expenses
**POST** `/api/expenses/export`

```bash
curl -X POST http://localhost:8080/api/expenses/export \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "line_u123456789",
    "format": "csv",
    "from": "2024-01-01T00:00:00Z",
    "to": "2024-01-31T23:59:59Z"
  }' \
  -o expenses.csv
```

Supported formats: `csv`, `json`

### Budget Management

#### Get Budget Status
**GET** `/api/budget/status`

```bash
curl "http://localhost:8080/api/budget/status?user_id=line_u123456789"
```

#### Compare to Budget
**GET** `/api/budget/compare`

```bash
curl "http://localhost:8080/api/budget/compare?user_id=line_u123456789"
```

### Recurring Expenses

#### Create Recurring Expense
**POST** `/api/recurring`

```bash
curl -X POST http://localhost:8080/api/recurring \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "line_u123456789",
    "description": "Monthly subscription",
    "amount": 99.99,
    "frequency": "monthly",
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-12-31T23:59:59Z"
  }'
```

Supported frequencies: `daily`, `weekly`, `monthly`, `yearly`

#### List Recurring Expenses
**GET** `/api/recurring?user_id=line_u123456789`

#### Update Recurring Expense
**PUT** `/api/recurring/{recurring_id}`

#### Delete Recurring Expense
**DELETE** `/api/recurring/{recurring_id}`

### Notifications

#### List Notifications
**GET** `/api/notifications?user_id=line_u123456789`

#### Create Notification
**POST** `/api/notifications`

```bash
curl -X POST http://localhost:8080/api/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "line_u123456789",
    "message": "Budget exceeded",
    "type": "alert"
  }'
```

Supported types: `info`, `warning`, `alert`

#### Mark as Read
**PUT** `/api/notifications/{notification_id}/read`

### Archive Management

#### Create Archive
**POST** `/api/archive`

```bash
curl -X POST http://localhost:8080/api/archive \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "line_u123456789",
    "name": "Year 2023",
    "from": "2023-01-01T00:00:00Z",
    "to": "2023-12-31T23:59:59Z"
  }'
```

#### List Archives
**GET** `/api/archive?user_id=line_u123456789`

### Health Check

#### Health Status
**GET** `/health`

```bash
curl http://localhost:8080/health
```

**Response** (200 OK):
```json
{
  "status": "healthy",
  "timestamp": "2024-01-18T10:30:00Z"
}
```

## Response Format

### Success Response
```json
{
  "status": "success",
  "message": "Operation completed successfully",
  "data": {}
}
```

### Error Response
```json
{
  "status": "error",
  "error": "Description of the error"
}
```

## HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | OK - Request successful |
| 201 | Created - Resource created |
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Missing or invalid API key |
| 404 | Not Found - Resource not found |
| 500 | Internal Server Error |

## Supported Messenger Platforms

- `line` - LINE Messaging API
- `telegram` - Telegram Bot API
- `slack` - Slack API
- `teams` - Microsoft Teams API
- `discord` - Discord API
- `whatsapp` - WhatsApp Business API

## Data Types

### User
```json
{
  "user_id": "string",
  "messenger_type": "string",
  "created_at": "ISO 8601 timestamp"
}
```

### Expense
```json
{
  "id": "string",
  "user_id": "string",
  "description": "string",
  "amount": "number",
  "category_id": "string",
  "expense_date": "ISO 8601 timestamp",
  "created_at": "ISO 8601 timestamp",
  "updated_at": "ISO 8601 timestamp"
}
```

### Category
```json
{
  "id": "string",
  "user_id": "string",
  "name": "string",
  "is_default": "boolean",
  "created_at": "ISO 8601 timestamp"
}
```

## Rate Limiting

Currently, rate limiting is not enforced. This may be added in future versions.

## Error Handling

The API uses standard HTTP status codes and returns error messages in JSON format:

```json
{
  "status": "error",
  "error": "Missing required fields: user_id and messenger_type"
}
```

## Examples

### Complete Workflow

1. **Auto-signup user**
```bash
curl -X POST http://localhost:8080/api/users/auto-signup \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user123", "messenger_type": "line"}'
```

2. **Parse expenses**
```bash
curl -X POST http://localhost:8080/api/expenses/parse \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user123", "text": "breakfast $20 lunch $30"}'
```

3. **Create expenses**
```bash
curl -X POST http://localhost:8080/api/expenses \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user123", "description": "breakfast", "amount": 20}'
```

4. **List expenses**
```bash
curl "http://localhost:8080/api/expenses?user_id=user123"
```

5. **Generate report**
```bash
curl -X POST http://localhost:8080/api/reports/generate \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user123", "from": "2024-01-01T00:00:00Z", "to": "2024-01-31T23:59:59Z"}'
```

## Tools for Testing

- **Swagger UI**: Import `openapi.yaml` into [swagger.io/swagger-ui](https://editor.swagger.io)
- **Postman**: Import `openapi.yaml` directly
- **cURL**: Use examples above
- **REST Client Extensions**: VSCode REST Client, Insomnia, ThunderClient

## OpenAPI Specification

The complete API specification is available in `openapi.yaml` in the root directory. It follows OpenAPI 3.0.0 standard and can be used to:

- Generate client SDKs
- Generate server stubs
- Validate requests/responses
- Generate interactive documentation
- Generate automated tests

## Support

For issues, questions, or feature requests, please refer to the [GitHub repository](https://github.com/riverlin/aiexpense).
