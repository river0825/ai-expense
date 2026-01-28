# Design: Payment Account Detection

## Backend Design

### Domain Model
1.  **Modify `Expense` struct**: Add `PaymentMethod` (string).
2.  **Modify `ParsedExpense` struct**: Add `PaymentMethod` (string).

### AI Service (`GeminiAI`)
*   Update the system prompt in `ParseExpense` to instruct the model to extract "payment_method" or "account".
*   Instruction: "Extract payment method if mentioned (e.g., 'Cash', 'Credit Card', 'Bank'). If not mentioned, return null or empty."
*   Update the parsing logic to populate `ParsedExpense.PaymentMethod`.
*   Handle the default "Cash" logic in the UseCase or AI service. Best to do it in UseCase if null.

### Database
*   **Migration**: `ALTER TABLE expenses ADD COLUMN payment_method TEXT DEFAULT 'Cash';` (Postgres).
*   **Repository**: Update `Create` and `Get` methods in `postgresql` repository to map this column. (Note: SQLite is no longer supported).

## Frontend Design

### Design System
- **Theme**: Dark Mode (Background: #0F172A, Primary: #F59E0B).
- **Typography**: IBM Plex Sans.
- **Components**: Mobile-first minimalist cards for expense items. Payment method displayed as a distinct badge/tag.

### Model
*   Update `Expense` interface in `dashboard/src/domain/models/Expense.ts`.

### UI
*   Update `ExpenseList` to show the payment method with the new design system.
*   Update `ExpenseForm` (if manual entry exists) to allow selecting/inputting payment method.
*   **Mobile Optimization**: Ensure touch targets are >44px and layout is single-column on small screens.

## API
*   The API response for `GET /expenses` will include the new field.
