import { test, expect } from '@playwright/test';

test.describe('Admin Panel', () => {
  test.describe('Login Flow', () => {
    test('should redirect to login when not authenticated', async ({ page }) => {
      await page.goto('/dashboard');
      await expect(page).toHaveURL(/.*login/);
    });

    test('should display login form', async ({ page }) => {
      await page.goto('/login');
      await expect(page.getByLabel(/username/i)).toBeVisible();
      await expect(page.getByLabel(/password/i)).toBeVisible();
      await expect(page.getByRole('button', { name: /sign in/i })).toBeVisible();
    });

    test('should login with valid credentials', async ({ page }) => {
      await page.goto('/login');
      await page.getByLabel(/username/i).fill('admin');
      await page.getByLabel(/password/i).fill('admin123');
      await page.getByRole('button', { name: /sign in/i }).click();
      
      await expect(page).toHaveURL('/dashboard');
    });
  });

  test.describe('Dashboard', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/login');
      await page.getByLabel(/username/i).fill('admin');
      await page.getByLabel(/password/i).fill('admin123');
      await page.getByRole('button', { name: /sign in/i }).click();
      await expect(page).toHaveURL('/dashboard');
    });

    test('should display KPI cards', async ({ page }) => {
      await expect(page.getByText('MRR')).toBeVisible();
      await expect(page.getByText('NRR')).toBeVisible();
      await expect(page.getByText('GRR')).toBeVisible();
      await expect(page.getByText('Churn Rate')).toBeVisible();
    });

    test('should display period filter buttons', async ({ page }) => {
      await expect(page.getByTestId('period-filter-7d')).toBeVisible();
      await expect(page.getByTestId('period-filter-30d')).toBeVisible();
      await expect(page.getByTestId('period-filter-90d')).toBeVisible();
    });

    test('should change period filter', async ({ page }) => {
      const filter30d = page.getByTestId('period-filter-30d');
      const filter90d = page.getByTestId('period-filter-90d');
      
      await filter90d.click();
      
      await expect(filter90d).toHaveClass(/bg-background/);
    });

    test('should display at-risk accounts table', async ({ page }) => {
      await expect(page.getByText('At-Risk Accounts')).toBeVisible();
      await expect(page.getByTestId('at-risk-table')).toBeVisible();
    });

    test('should display revenue trend chart', async ({ page }) => {
      await expect(page.getByText('Revenue Trend')).toBeVisible();
      await expect(page.getByTestId('trend-chart')).toBeVisible();
    });
  });

  test.describe('Responsive Design', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/login');
      await page.getByLabel(/username/i).fill('admin');
      await page.getByLabel(/password/i).fill('admin123');
      await page.getByRole('button', { name: /sign in/i }).click();
      await expect(page).toHaveURL('/dashboard');
    });

    test('should adapt to mobile viewport', async ({ page }) => {
      await page.setViewportSize({ width: 390, height: 844 });
      
      await expect(page.getByText('Revenue Command Center')).toBeVisible();
      await expect(page.getByTestId('period-filter-30d')).toBeVisible();
    });

    test('should display properly on desktop', async ({ page }) => {
      await page.setViewportSize({ width: 1440, height: 900 });
      
      await expect(page.getByText('Revenue Command Center')).toBeVisible();
      const kpiGrid = page.locator('.grid');
      await expect(kpiGrid).toBeVisible();
    });
  });
});
