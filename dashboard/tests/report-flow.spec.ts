import { test, expect } from '@playwright/test';

test.describe('User Report Dashboard Flow', () => {
  // Mock API responses
  const mockReportData = {
    status: 'success',
    data: {
      total_expenses: 1250.50,
      transaction_count: 15,
      average_expense: 83.37,
      highest_expense: 250.00,
      category_breakdown: [
        { category: 'Food', total: 450.00, count: 5, percentage: 36 },
        { category: 'Transport', total: 300.50, count: 4, percentage: 24 },
        { category: 'Utilities', total: 500.00, count: 6, percentage: 40 }
      ],
      top_expenses: [
        { id: '1', description: 'Groceries', amount: 150.00, category: 'Food', date: new Date().toISOString() },
        { id: '2', description: 'Electric Bill', amount: 200.00, category: 'Utilities', date: new Date().toISOString() }
      ],
      period: 'Last 30 Days'
    }
  };

  test('full report flow from short link to dashboard', async ({ page }) => {
    // 1. Simulate Short Link Redirection
    // Since we can't easily spin up the Go backend in this environment, 
    // we'll go directly to the target URL that the backend WOULD redirect to.
    const validToken = 'mock_valid_jwt_token';
    const targetUrl = `/reports?token=${validToken}`;

    // 2. Mock the API call that the dashboard will make
    await page.route('**/api/reports/summary*', async route => {
      // Verify token is passed in Authorization header or query param
      const headers = route.request().headers();
      const url = route.request().url();
      
      if (url.includes(`token=${validToken}`) || (headers['authorization'] && headers['authorization'].includes(validToken))) {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(mockReportData)
        });
      } else {
        await route.fulfill({
          status: 401,
          contentType: 'application/json',
          body: JSON.stringify({ status: 'error', error: 'Unauthorized' })
        });
      }
    });

    // 3. Navigate to the dashboard (simulating the redirect landing)
    await page.goto(targetUrl, { waitUntil: 'networkidle' });

    // 4. Verify Dashboard Loads
    await expect(page).toHaveTitle(/Dashboard/i);
    
    // 5. Verify Metrics Display
    // Note: Recharts and some dynamic content might take time or render differently in headless
    // We check for text existence more loosely if strict locators fail
    await expect(page.getByText('$1250.50')).toBeVisible({ timeout: 10000 }); // Total Expense
    await expect(page.getByText('15', { exact: true })).toBeVisible(); // Transaction Count
    
    // 6. Verify Chart & List Elements
    await expect(page.getByText('Spending by Category')).toBeVisible();
    await expect(page.getByText('Food')).toBeVisible(); // Chart legend/tooltip logic might need specific selectors
    
    await expect(page.getByText('Groceries')).toBeVisible(); // Expense List item
    await expect(page.getByText('-$150.00')).toBeVisible(); // Formatted amount

    // 7. Test Date Range Interaction (Basic UI check)
    const datePicker = page.locator('button#date'); // Assuming ID or we find by role
    await expect(datePicker).toBeVisible();
    await datePicker.click();
    await expect(page.locator('.rdp')).toBeVisible(); // react-day-picker uses .rdp class usually, or check for calendar role
  });

  test('invalid token shows error', async ({ page }) => {
    await page.route('**/api/reports/summary*', async route => {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({ status: 'error', error: 'Invalid or expired token' })
      });
    });

    await page.goto('/reports?token=invalid_token');
    await expect(page.getByText('Access Denied')).toBeVisible();
    await expect(page.getByText('Invalid or expired token')).toBeVisible();
  });
});
