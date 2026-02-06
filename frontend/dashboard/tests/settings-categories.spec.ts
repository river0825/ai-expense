import { test, expect } from '@playwright/test';

test.describe('Settings - Category Management', () => {
  const testUser = 'test_user';
  const settingsUrl = `/en/dashboard/settings?token=${testUser}`;

  test.beforeEach(async ({ page }) => {
    await page.goto(`/en/dashboard/settings?token=${testUser}`, { waitUntil: 'networkidle' });
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    
    await page.waitForSelector('h2:has-text("Category Management")', { timeout: 15000 }).catch(() => {});
  });

  test('displays categories with descriptions', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Category Management' })).toBeVisible({ timeout: 10000 });
    await page.waitForTimeout(500);
    
    const addButton = page.getByTestId('add-category-button');
    await expect(addButton).toBeVisible({ timeout: 5000 });
  });

  test('add new category with description', async ({ page }) => {
    const addButton = page.getByTestId('add-category-button');
    await addButton.click();
    await page.waitForTimeout(300);
    
    const nameInput = page.getByTestId('category-name-input');
    await nameInput.fill('Test Category New');
    
    const descInput = page.getByTestId('category-description-input');
    await descInput.fill('Test description for new category');
    
    const saveButton = page.getByTestId('save-category-button');
    await saveButton.click();
    await page.waitForTimeout(500);
    
    await expect(page.getByText('Test Category New')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Test description for new category')).toBeVisible();
    
    await page.goto(settingsUrl, { waitUntil: 'networkidle' });
    await page.waitForTimeout(500);
    
    const categoryRow = page.locator('text=Test Category New').first();
    const deleteButton = categoryRow.locator('..').getByTestId('delete-category-button');
    if (await deleteButton.isVisible()) {
      await deleteButton.click();
      const confirmButton = page.getByTestId('confirm-delete-button');
      await confirmButton.click();
      await page.waitForTimeout(500);
    }
  });

  test('edit category name and description', async ({ page }) => {
    const addButton = page.getByTestId('add-category-button');
    await addButton.click();
    await page.waitForTimeout(300);
    
    const nameInput = page.getByTestId('category-name-input');
    await nameInput.fill('Original Category');
    
    const descInput = page.getByTestId('category-description-input');
    await descInput.fill('Original description');
    
    const saveButton = page.getByTestId('save-category-button');
    await saveButton.click();
    
    await page.waitForTimeout(500);
    await expect(page.getByText('Original Category')).toBeVisible();
    
    const categoryRow = page.locator('text=Original Category').first();
    const editButton = categoryRow.locator('..').getByTestId('edit-category-button');
    await editButton.click();
    
    await page.waitForTimeout(300);
    
    const editNameInput = page.getByTestId('category-name-input');
    await editNameInput.clear();
    await editNameInput.fill('Updated Category');
    
    const editDescInput = page.getByTestId('category-description-input');
    await editDescInput.clear();
    await editDescInput.fill('Updated description');
    
    const saveEditButton = page.getByTestId('save-edit-button');
    await saveEditButton.click();
    
    await page.waitForTimeout(500);
    
    await expect(page.getByText('Updated Category')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Updated description')).toBeVisible();
    
    await page.goto(settingsUrl, { waitUntil: 'networkidle' });
    await page.waitForTimeout(500);
    
    const updatedRow = page.locator('text=Updated Category').first();
    const deleteBtn = updatedRow.locator('..').getByTestId('delete-category-button');
    if (await deleteBtn.isVisible()) {
      await deleteBtn.click();
      const confirmBtn = page.getByTestId('confirm-delete-button');
      await confirmBtn.click();
      await page.waitForTimeout(500);
    }
  });

  test('merge categories when editing to duplicate name', async ({ page }) => {
    const addButton = page.getByTestId('add-category-button');
    await addButton.click();
    await page.waitForTimeout(300);
    
    const nameInput = page.getByTestId('category-name-input');
    await nameInput.fill('Source Category');
    
    const descInput = page.getByTestId('category-description-input');
    await descInput.fill('Category to be merged');
    
    let saveButton = page.getByTestId('save-category-button');
    await saveButton.click();
    await page.waitForTimeout(500);
    
    await addButton.click();
    await page.waitForTimeout(300);
    
    const nameInput2 = page.getByTestId('category-name-input');
    await nameInput2.fill('Target Category');
    
    const descInput2 = page.getByTestId('category-description-input');
    await descInput2.fill('Target category');
    
    saveButton = page.getByTestId('save-category-button');
    await saveButton.click();
    await page.waitForTimeout(500);
    
    const sourceRow = page.locator('text=Source Category').first();
    const editButton = sourceRow.locator('..').getByTestId('edit-category-button');
    await editButton.click();
    
    await page.waitForTimeout(300);
    
    const editNameInput = page.getByTestId('category-name-input');
    await editNameInput.clear();
    await editNameInput.fill('Target Category');
    
    const saveEditButton = page.getByTestId('save-edit-button');
    await saveEditButton.click();
    
    await expect(page.getByText(/merge|Merge/i)).toBeVisible({ timeout: 5000 });
    
    const mergeConfirmButton = page.getByTestId('merge-confirm-button');
    await mergeConfirmButton.click();
    
    await page.waitForTimeout(500);
    
    const sourceElements = page.locator('text=Source Category');
    expect(await sourceElements.count()).toBe(0);
    
    await expect(page.getByText('Target Category')).toBeVisible();
    
    await page.goto(settingsUrl, { waitUntil: 'networkidle' });
    await page.waitForTimeout(500);
    
    const targetRow = page.locator('text=Target Category').first();
    const deleteBtn = targetRow.locator('..').getByTestId('delete-category-button');
    if (await deleteBtn.isVisible()) {
      await deleteBtn.click();
      const confirmBtn = page.getByTestId('confirm-delete-button');
      await confirmBtn.click();
      await page.waitForTimeout(500);
    }
  });

  test('delete category with confirmation', async ({ page }) => {
    const addButton = page.getByTestId('add-category-button');
    await addButton.click();
    await page.waitForTimeout(300);
    
    const nameInput = page.getByTestId('category-name-input');
    await nameInput.fill('Delete Test Category');
    
    const descInput = page.getByTestId('category-description-input');
    await descInput.fill('Will be deleted');
    
    const saveButton = page.getByTestId('save-category-button');
    await saveButton.click();
    
    await page.waitForTimeout(500);
    await expect(page.getByText('Delete Test Category')).toBeVisible();
    
    const categoryRow = page.locator('text=Delete Test Category').first();
    const deleteButton = categoryRow.locator('..').getByTestId('delete-category-button');
    await deleteButton.click();
    
    const confirmButton = page.getByTestId('confirm-delete-button');
    await expect(confirmButton).toBeVisible();
    await confirmButton.click();
    
    await page.waitForTimeout(500);
    
    const deletedElements = page.locator('text=Delete Test Category');
    expect(await deletedElements.count()).toBe(0);
  });

  test('cancel delete keeps category', async ({ page }) => {
    const addButton = page.getByTestId('add-category-button');
    await addButton.click();
    await page.waitForTimeout(300);
    
    const nameInput = page.getByTestId('category-name-input');
    await nameInput.fill('Keep This Category');
    
    const saveButton = page.getByTestId('save-category-button');
    await saveButton.click();
    
    await page.waitForTimeout(500);
    await expect(page.getByText('Keep This Category')).toBeVisible();
    
    const categoryRow = page.locator('text=Keep This Category').first();
    const deleteButton = categoryRow.locator('..').getByTestId('delete-category-button');
    await deleteButton.click();
    
    const cancelButton = page.getByTestId('cancel-delete-button');
    await cancelButton.click();
    
    await page.waitForTimeout(300);
    await expect(page.getByText('Keep This Category')).toBeVisible();
    
    await page.goto(settingsUrl, { waitUntil: 'networkidle' });
    await page.waitForTimeout(500);
    
    const row = page.locator('text=Keep This Category').first();
    const deleteBtn = row.locator('..').getByTestId('delete-category-button');
    if (await deleteBtn.isVisible()) {
      await deleteBtn.click();
      const confirmBtn = page.getByTestId('confirm-delete-button');
      await confirmBtn.click();
      await page.waitForTimeout(500);
    }
  });

  test('default categories are protected from editing', async ({ page }) => {
    const defaultCategorySection = page.getByText(/Default/i).first();
    await expect(defaultCategorySection).toBeVisible();
    
    const defaultRow = defaultCategorySection.locator('..');
    
    const editButtons = defaultRow.getByTestId('edit-category-button');
    const editCount = await editButtons.count();
    expect(editCount).toBe(0);
    
    const deleteButtons = defaultRow.getByTestId('delete-category-button');
    const deleteCount = await deleteButtons.count();
    expect(deleteCount).toBe(0);
  });

  test('cancel edit keeps original values', async ({ page }) => {
    const addButton = page.getByTestId('add-category-button');
    await addButton.click();
    await page.waitForTimeout(300);
    
    const nameInput = page.getByTestId('category-name-input');
    await nameInput.fill('Cancel Edit Test');
    
    const descInput = page.getByTestId('category-description-input');
    await descInput.fill('Original values');
    
    const saveButton = page.getByTestId('save-category-button');
    await saveButton.click();
    
    await page.waitForTimeout(500);
    
    const categoryRow = page.locator('text=Cancel Edit Test').first();
    const editButton = categoryRow.locator('..').getByTestId('edit-category-button');
    await editButton.click();
    
    await page.waitForTimeout(300);
    
    const editNameInput = page.getByTestId('category-name-input');
    await editNameInput.clear();
    await editNameInput.fill('Changed Name');
    
    const cancelEditButton = page.getByTestId('cancel-edit-button');
    if (await cancelEditButton.isVisible()) {
      await cancelEditButton.click();
    }
    
    await page.waitForTimeout(300);
    await expect(page.getByText('Cancel Edit Test')).toBeVisible();
    
    await page.goto(settingsUrl, { waitUntil: 'networkidle' });
    await page.waitForTimeout(500);
    
    const row = page.locator('text=Cancel Edit Test').first();
    const deleteBtn = row.locator('..').getByTestId('delete-category-button');
    if (await deleteBtn.isVisible()) {
      await deleteBtn.click();
      const confirmBtn = page.getByTestId('confirm-delete-button');
      await confirmBtn.click();
      await page.waitForTimeout(500);
    }
  });
});
