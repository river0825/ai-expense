import { test, expect } from '@playwright/test';

test('dashboard flow', async ({ page }) => {
  await page.goto('/');

  await expect(page.getByRole('heading', { name: 'AIExpense' })).toBeVisible();
  await expect(page.getByText('Metrics Dashboard')).toBeVisible();
  
  const apiKeyInput = page.getByPlaceholder('Enter your admin API key');
  await expect(apiKeyInput).toBeVisible();
  await apiKeyInput.fill('admin_key');
  
  const submitButton = page.getByRole('button', { name: 'View Metrics' });
  await expect(submitButton).toBeVisible();
  await submitButton.click();

  await expect(page.getByText('Loading metrics...')).toBeHidden({ timeout: 10000 });
  
  await expect(page.getByRole('banner')).toBeVisible();
  await expect(page.getByText('AIExpense Metrics')).toBeVisible();
  
  await expect(page.getByRole('button', { name: 'Logout' })).toBeVisible();
});
