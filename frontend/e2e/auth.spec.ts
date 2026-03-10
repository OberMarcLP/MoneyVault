import { test, expect } from '@playwright/test';
import { uniqueEmail, registerUser, loginUser, TEST_PASSWORD } from './helpers';

test.describe('Authentication', () => {
  test('register new account redirects to login', async ({ page }) => {
    const email = uniqueEmail('reg');
    await registerUser(page, email);
    expect(page.url()).toContain('/login');
  });

  test('login with valid credentials redirects to dashboard', async ({ page }) => {
    const email = uniqueEmail('login');
    await registerUser(page, email);
    await loginUser(page, email);
    await expect(page).not.toHaveURL(/\/login/);
  });

  test('login with wrong password shows error', async ({ page }) => {
    const email = uniqueEmail('wrong');
    await registerUser(page, email);
    await page.goto('/login');
    await page.getByPlaceholder('you@example.com').fill(email);
    await page.getByPlaceholder('Enter your password').fill('WrongPass999!');
    await page.getByRole('button', { name: 'Sign in', exact: true }).click();
    await expect(page.locator('.text-destructive')).toBeVisible({ timeout: 5000 });
  });

  test('login with non-existent email shows error', async ({ page }) => {
    await page.goto('/login');
    await page.getByPlaceholder('you@example.com').fill('nobody@nowhere.test');
    await page.getByPlaceholder('Enter your password').fill(TEST_PASSWORD);
    await page.getByRole('button', { name: 'Sign in', exact: true }).click();
    await expect(page.locator('.text-destructive')).toBeVisible({ timeout: 5000 });
  });

  test('register with weak password keeps submit disabled', async ({ page }) => {
    await page.goto('/register');
    await page.getByPlaceholder('you@example.com').fill(uniqueEmail('weak'));
    await page.getByPlaceholder('Create a strong password').fill('short');
    await page.getByPlaceholder('Repeat your password').fill('short');
    await expect(page.getByRole('button', { name: 'Create account' })).toBeDisabled();
  });

  test('register with mismatched passwords shows error', async ({ page }) => {
    await page.goto('/register');
    await page.getByPlaceholder('you@example.com').fill(uniqueEmail('mismatch'));
    await page.getByPlaceholder('Create a strong password').fill(TEST_PASSWORD);
    await page.getByPlaceholder('Repeat your password').fill('Different1!');
    // Trigger blur to show error
    await page.getByPlaceholder('Repeat your password').blur();
    await expect(page.getByText('do not match')).toBeVisible({ timeout: 3000 });
  });

  test('unauthenticated access to /accounts redirects to login', async ({ page }) => {
    await page.goto('/accounts');
    await page.waitForURL('**/login', { timeout: 5000 });
    expect(page.url()).toContain('/login');
  });
});
