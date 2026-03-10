import { test, expect } from '@playwright/test';
import { registerUser, loginUser, uniqueEmail, createAccount, navigateTo, TEST_PASSWORD } from './helpers';

test.describe('Transactions', () => {
  let email: string;

  test.beforeAll(async ({ browser }) => {
    email = uniqueEmail('tx');
    const page = await browser.newPage();
    await registerUser(page, email);
    await loginUser(page, email);
    await createAccount(page, 'Test Account');
    await page.close();
  });

  test.beforeEach(async ({ page }) => {
    await loginUser(page, email);
  });

  async function addTransaction(page: import('@playwright/test').Page, opts: {
    type?: string; amount: string; description: string; date?: string;
  }) {
    await page.getByRole('button', { name: 'Add Transaction' }).click();
    await expect(page.getByRole('heading', { name: 'Add Transaction' })).toBeVisible();

    if (opts.type && opts.type !== 'expense') {
      await page.locator('form select').first().selectOption(opts.type);
    }
    // Wait for accounts to load in the select, then pick the test account
    const accountSelect = page.locator('form select').nth(1);
    await expect(accountSelect).toContainText('Test Account', { timeout: 5000 });
    await accountSelect.selectOption({ label: 'Test Account' });
    await page.locator('input[type="date"]').fill(opts.date ?? '2025-06-15');
    await page.locator('input[type="number"]').fill(opts.amount);
    await page.getByPlaceholder('What was this for?').fill(opts.description);
    // Use form submit button specifically (page also has an "Add Transaction" button)
    await page.locator('form button[type="submit"]').click();
    await expect(page.getByText(opts.description)).toBeVisible({ timeout: 5000 });
  }

  test('create expense transaction appears in list', async ({ page }) => {
    await navigateTo(page, 'Transactions');
    await addTransaction(page, { amount: '42.50', description: 'E2E coffee' });
  });

  test('filter by type shows correct transactions', async ({ page }) => {
    await navigateTo(page, 'Transactions');
    await addTransaction(page, { amount: '100', description: 'Expense item' });
    await addTransaction(page, { type: 'income', amount: '500', description: 'Income item', date: '2025-06-16' });

    // Filter by income
    await page.locator('.flex.flex-wrap select').first().selectOption('income');
    await expect(page.getByText('Income item')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Expense item')).not.toBeVisible();

    // Filter by expense
    await page.locator('.flex.flex-wrap select').first().selectOption('expense');
    await expect(page.getByText('Expense item')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Income item')).not.toBeVisible();
  });

  test('search by description filters transactions', async ({ page }) => {
    await navigateTo(page, 'Transactions');
    await addTransaction(page, { amount: '25', description: 'Unique alpha' });
    await addTransaction(page, { amount: '35', description: 'Unique beta', date: '2025-06-16' });

    await page.getByPlaceholder('Search transactions...').fill('alpha');
    await expect(page.getByText('Unique alpha')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Unique beta')).not.toBeVisible();
  });

  test('delete transaction with confirmation removes it', async ({ page }) => {
    await navigateTo(page, 'Transactions');
    await addTransaction(page, { amount: '15', description: 'Delete me tx' });

    // Click the delete button in the specific "Delete me tx" row
    const txRow = page.getByText('Delete me tx', { exact: true })
      .locator('xpath=ancestor::div[contains(@class, "justify-between")][1]');
    await txRow.locator('button').last().click();
    await expect(page.getByText('Delete Transaction')).toBeVisible();
    // Target the ConfirmDialog's destructive button
    await page.locator('button.bg-destructive').click();
    await expect(page.getByText('Delete me tx')).not.toBeVisible({ timeout: 5000 });
  });
});
