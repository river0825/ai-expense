# OpenAPI Specification Summary

## Files Generated

### 1. `openapi.yaml` (Root Directory)
- **Format**: OpenAPI 3.0.0 YAML
- **Size**: ~1500 lines
- **Location**: `/openapi.yaml`
- **Purpose**: Machine-readable API specification

### 2. `docs/API.md` (This Directory)
- **Format**: Markdown
- **Purpose**: Human-readable API documentation with examples
- **Location**: `/docs/API.md`

## Specification Overview

### API Title
**AIExpense API v1.0.0**

Conversational Expense Tracking System - A REST API for managing expenses through natural language conversation

### Server Endpoints
- **Development**: `http://localhost:8080`
- **Production**: `https://api.aiexpense.app`

## API Endpoints Covered (50+)

### Users (1)
- ✅ POST `/api/users/auto-signup` - Auto-signup user

### Expenses (7)
- ✅ POST `/api/expenses` - Create expense
- ✅ GET `/api/expenses` - List expenses
- ✅ PUT `/api/expenses/{id}` - Update expense
- ✅ DELETE `/api/expenses/{id}` - Delete expense
- ✅ GET `/api/expenses/search` - Search expenses
- ✅ GET `/api/expenses/filter` - Filter expenses
- ✅ POST `/api/expenses/parse` - Parse natural language

### Categories (4)
- ✅ POST `/api/categories` - Create category
- ✅ GET `/api/categories` - List categories
- ✅ PUT `/api/categories/{id}` - Update category
- ✅ DELETE `/api/categories/{id}` - Delete category

### Metrics (3)
- ✅ GET `/api/metrics/dau` - Daily active users
- ✅ GET `/api/metrics/expenses-summary` - Expense summary
- ✅ GET `/api/metrics/growth` - Growth metrics

### Reports & Export (2)
- ✅ POST `/api/reports/generate` - Generate report
- ✅ POST `/api/expenses/export` - Export expenses

### Budget (2)
- ✅ GET `/api/budget/status` - Budget status
- ✅ GET `/api/budget/compare` - Compare to budget

### Recurring (5)
- ✅ POST `/api/recurring` - Create recurring expense
- ✅ GET `/api/recurring` - List recurring expenses
- ✅ PUT `/api/recurring/{id}` - Update recurring expense
- ✅ DELETE `/api/recurring/{id}` - Delete recurring expense
- ✅ GET `/api/recurring/upcoming` - Get upcoming recurring

### Notifications (5)
- ✅ POST `/api/notifications` - Create notification
- ✅ GET `/api/notifications` - List notifications
- ✅ PUT `/api/notifications/{id}/read` - Mark as read
- ✅ PUT `/api/notifications/read-all` - Mark all as read
- ✅ DELETE `/api/notifications/{id}` - Delete notification

### Archive (5)
- ✅ POST `/api/archive` - Create archive
- ✅ GET `/api/archive` - List archives
- ✅ GET `/api/archive/{id}/stats` - Archive stats
- ✅ GET `/api/archive/{id}/details` - Archive details
- ✅ PUT `/api/archive/{id}/restore` - Restore archive

### Health (1)
- ✅ GET `/health` - Health check

## Data Schemas Defined

### Request Schemas
- ✅ AutoSignupRequest
- ✅ ParseRequest
- ✅ CreateExpenseRequest
- ✅ UpdateExpenseRequest
- ✅ CreateCategoryRequest
- ✅ UpdateCategoryRequest
- ✅ GenerateReportRequest
- ✅ ExportRequest
- ✅ CreateRecurringRequest
- ✅ UpdateRecurringRequest
- ✅ CreateNotificationRequest
- ✅ CreateArchiveRequest

### Response Schemas
- ✅ SuccessResponse
- ✅ ErrorResponse
- ✅ Expense
- ✅ Category
- ✅ ParsedExpense

### Support Data Types
- ✅ User
- ✅ Expense
- ✅ Category
- ✅ Notification
- ✅ Archive

## Authentication

### Security Schemes
- **ApiKeyAuth**: Header-based API key authentication
  - Header: `X-API-Key`
  - Required for: Metrics endpoints
  - Type: Admin API key

## Supported Messenger Platforms

1. LINE (`line`)
2. Telegram (`telegram`)
3. Slack (`slack`)
4. Microsoft Teams (`teams`)
5. Discord (`discord`)
6. WhatsApp (`whatsapp`)

## HTTP Methods & Status Codes

### Methods
- ✅ GET - Retrieve data
- ✅ POST - Create data (201 Created)
- ✅ PUT - Update data
- ✅ DELETE - Delete data

### Status Codes
- ✅ 200 - OK
- ✅ 201 - Created
- ✅ 400 - Bad Request
- ✅ 401 - Unauthorized
- ✅ 404 - Not Found
- ✅ 500 - Internal Server Error

## Tags for Organization

The specification organizes endpoints into 10 logical tags:
1. Users
2. Expenses
3. Categories
4. Metrics
5. Reports
6. Budget
7. Recurring
8. Notifications
9. Archive
10. Health

## OpenAPI Standard Compliance

✅ **OpenAPI 3.0.0** compliant
- Machine-readable format
- Tool-compatible (Swagger UI, Postman, code generators)
- Comprehensive documentation
- Request/response schema validation

## How to Use the Specifications

### For Development
1. **Swagger UI**: Visit [https://editor.swagger.io](https://editor.swagger.io)
2. **Import**: File → Import YAML → Select `openapi.yaml`
3. **Test**: Try out endpoints directly in Swagger UI

### For Integration
1. **Postman**: Collections → Import → Select `openapi.yaml`
2. **Code Generation**: Use tools like:
   - OpenAPI Generator: Generate client libraries for any language
   - AutoRest: Generate SDKs from OpenAPI
   - Swagger Codegen

### For Documentation
- Read `docs/API.md` for human-readable examples
- Copy-paste cURL examples for quick testing
- Use interactive Swagger UI for exploration

### For Testing
- Generate test cases from schema
- Validate request/response against spec
- Automated API testing tools can use the spec

## Common Use Cases

### 1. Auto-signup and Create First Expense
```bash
# 1. Auto-signup
curl -X POST http://localhost:8080/api/users/auto-signup \
  -d '{"user_id": "user123", "messenger_type": "line"}'

# 2. Parse
curl -X POST http://localhost:8080/api/expenses/parse \
  -d '{"user_id": "user123", "text": "breakfast $20"}'

# 3. Create
curl -X POST http://localhost:8080/api/expenses \
  -d '{"user_id": "user123", "description": "breakfast", "amount": 20}'
```

### 2. Generate Monthly Report
```bash
curl -X POST http://localhost:8080/api/reports/generate \
  -d '{
    "user_id": "user123",
    "from": "2024-01-01T00:00:00Z",
    "to": "2024-01-31T23:59:59Z"
  }'
```

### 3. View Metrics (Admin)
```bash
curl http://localhost:8080/api/metrics/dau \
  -H "X-API-Key: admin-key-123"
```

## File Locations

```
aiexpense/
├── openapi.yaml                    # Main OpenAPI specification (YAML)
├── docs/
│   ├── API.md                     # Human-readable API documentation
│   └── OPENAPI_SPEC.md           # This file
└── ... (other project files)
```

## Next Steps

1. **Review**: Check `openapi.yaml` and `docs/API.md`
2. **Test**: Import into Swagger UI or Postman
3. **Generate**: Use for code generation or automation
4. **Implement**: Route implementations should match spec
5. **Validate**: Ensure handlers return correct status codes/schemas

## Tools & Resources

### Online Tools
- [Swagger Editor](https://editor.swagger.io) - Online OpenAPI editor
- [Swagger UI](https://swagger.io/tools/swagger-ui/) - Interactive API docs
- [OpenAPI Generator](https://openapi-generator.tech) - Generate code from spec

### Local Tools
- Postman: Import openapi.yaml directly
- IntelliJ IDEA: Built-in OpenAPI support
- VSCode: Extensions for OpenAPI editing
- Insomnia: OpenAPI import support

### Command Line
```bash
# Validate OpenAPI spec
npm install -g @openapitools/openapi-generator-cli
openapi-generator-cli validate -i openapi.yaml

# Generate client (Python example)
openapi-generator-cli generate \
  -i openapi.yaml \
  -g python \
  -o ./generated-client
```

## Support

For questions about the API specification, refer to:
- `docs/API.md` - API documentation
- `openapi.yaml` - Complete specification
- GitHub Issues - Report issues or request features
