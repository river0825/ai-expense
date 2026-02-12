import { expect, test } from '@playwright/test';
import path from 'path';

test('capture dashboard screenshot', async ({ page }) => {
  // Mock API responses to ensure valid data and no errors
  await page.route('**/api/reports/summary*', async route => {
    const json = {
      status: 'success',
      data: {
        total_spent: 3820.00,
        currency: 'USD',
        user_id: 'user_123',
        category_breakdown: [
          { category: 'Food', total: 1200, count: 15, percentage: 31.4 },
          { category: 'Transport', total: 800, count: 10, percentage: 20.9 },
          { category: 'Utilities', total: 1820, count: 5, percentage: 47.6 }
        ],
        daily_breakdown: [
          { date: '2023-10-01', total: 100, count: 2 },
          { date: '2023-10-02', total: 200, count: 3 },
          { date: '2023-10-03', total: 150, count: 1 }
        ],
        top_expenses: [
          {
            id: 'exp_1',
            description: 'Lunch at Cafe',
            amount: 25.50,
            category: 'Food',
            date: '2023-10-05',
            currency: 'USD',
            account: 'Credit Card'
          },
          {
            id: 'exp_2',
            description: 'Uber Ride',
            amount: 15.00,
            category: 'Transport',
            date: '2023-10-05',
            currency: 'USD',
            account: 'Credit Card'
          }
        ]
      }
    };
    await route.fulfill({ json });
  });

  await page.route('**/api/user*', async route => {
    await route.fulfill({ json: { status: 'success', data: { id: 'user_123', name: 'Demo User', email: 'demo@example.com' } } });
  });
  
  await page.route('**/api/currencies*', async route => {
     await route.fulfill({ json: { status: 'success', data: ['USD', 'EUR', 'TWD', 'JPY'] } });
  });

   await page.route('**/api/categories*', async route => {
     await route.fulfill({ json: { status: 'success', data: ['Food', 'Transport', 'Utilities', 'Entropy'] } });
  });


  // Navigate to dashboard with token to bypass client-side check
  await page.goto('/en/dashboard?token=demo_token');

  // Handle Login if present (legacy or different route?)
  const apiKeyInput = page.getByPlaceholder('Enter your admin API key');
  if (await apiKeyInput.isVisible()) {
    await apiKeyInput.fill('demo_key');
    await page.getByRole('button', { name: 'View Metrics' }).click();
  }

  // Wait for dashboard to load
  await expect(page.getByText('Loading metrics...')).toBeHidden({ timeout: 15000 });
  await expect(page.getByRole('heading', { name: 'My Expenses' })).toBeVisible();

  // Wait for animations (charts etc)
  await page.waitForTimeout(2000);

  // Take Screenshot
  // Take Screenshot
  const screenshotPath = path.resolve(process.cwd(), '../landing/assets/dashboard-mockup.png');
  await page.screenshot({ path: screenshotPath, fullPage: true });
  console.log(`Screenshot saved to ${screenshotPath}`);
});
