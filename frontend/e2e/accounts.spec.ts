import { test, expect } from '@playwright/test';
import { registerAndLogin, createAccount } from './helpers';

test.describe('Accounts', () => {
  test.beforeEach(async ({ page }) => {
    await registerAndLogin(page);
  });

  test('create account appears in list', async ({ page }) => {
    await createAccount(page, 'E2E Checking');
    await expect(page.getByText('E2E Checking')).toBeVisible();
  });

  test('edit account name updates in list', async ({ page }) => {
    await createAccount(page, 'Before Edit');
    await page.getByRole('button', { name: 'Edit' }).first().click();
    await expect(page.getByRole('heading', { name: 'Edit Account' })).toBeVisible();
    await page.getByPlaceholder('e.g. Main Checking').fill('After Edit');
    await page.getByRole('button', { name: 'Update Account' }).click();
    await expect(page.getByText('After Edit')).toBeVisible({ timeout: 5000 });
  });

  test('delete with cancel keeps account', async ({ page }) => {
    await createAccount(page, 'Keep Me');
    await page.getByRole('button', { name: 'Delete' }).first().click();
    await expect(page.getByRole('heading', { name: 'Delete Account' })).toBeVisible();
    await page.getByRole('button', { name: 'Cancel' }).click();
    await expect(page.getByText('Keep Me')).toBeVisible();
  });

  test('delete with confirm removes account', async ({ page }) => {
    await createAccount(page, 'Remove Me');
    await expect(page.getByText('Remove Me')).toBeVisible();
    await page.getByRole('button', { name: 'Delete' }).first().click();
    await expect(page.getByRole('heading', { name: 'Delete Account' })).toBeVisible();
    // Target the ConfirmDialog's destructive button (not the card's outline Delete button)
    await page.locator('button.bg-destructive').click();
    // Use heading role to avoid matching the ConfirmDialog message text
    await expect(page.getByRole('heading', { name: 'Remove Me' })).not.toBeVisible({ timeout: 5000 });
  });
});
