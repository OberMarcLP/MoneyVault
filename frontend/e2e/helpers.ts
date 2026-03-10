import { type Page, expect } from '@playwright/test';

export const TEST_PASSWORD = 'E2eTest@2026!';

export function uniqueEmail(prefix = 'e2e'): string {
  return `${prefix}+${Date.now()}@e2e.test`;
}

export async function registerUser(page: Page, email: string, password = TEST_PASSWORD) {
  await page.goto('/register');
  await page.getByPlaceholder('you@example.com').fill(email);
  await page.getByPlaceholder('Create a strong password').fill(password);
  await page.getByPlaceholder('Repeat your password').fill(password);
  await expect(page.getByRole('button', { name: 'Create account' })).toBeEnabled({ timeout: 3000 });
  await page.getByRole('button', { name: 'Create account' }).click();
  await page.waitForURL('**/login', { timeout: 10_000 });
}

export async function loginUser(page: Page, email: string, password = TEST_PASSWORD) {
  await page.goto('/login');
  await page.getByPlaceholder('you@example.com').fill(email);
  await page.getByPlaceholder('Enter your password').fill(password);
  await page.getByRole('button', { name: 'Sign in', exact: true }).click();
  await page.waitForURL('**/', { timeout: 10_000 });
}

export async function registerAndLogin(page: Page): Promise<string> {
  const email = uniqueEmail();
  await registerUser(page, email);
  await loginUser(page, email);
  return email;
}

/** Navigate to a page via sidebar link (avoids full page reload). */
export async function navigateTo(page: Page, name: string) {
  await page.getByRole('link', { name: new RegExp(name, 'i') }).click();
  await page.waitForURL(`**/${name.toLowerCase()}`, { timeout: 5000 });
}

export async function createAccount(page: Page, name: string, type = 'checking', balance = '1000') {
  await navigateTo(page, 'Accounts');
  // Two "Add Account" buttons when empty (header + empty state), use first
  await page.getByRole('button', { name: 'Add Account' }).first().click();
  // Wait for dialog heading
  await expect(page.getByRole('heading', { name: 'Create Account' })).toBeVisible();
  await page.getByPlaceholder('e.g. Main Checking').fill(name);
  if (type !== 'checking') {
    await page.locator('form select').first().selectOption(type);
  }
  await page.locator('input[type="number"]').fill(balance);
  // Submit button text is "Create Account"
  await page.getByRole('button', { name: 'Create Account' }).click();
  await expect(page.getByText(name)).toBeVisible({ timeout: 5000 });
}
