import { test, expect } from '@playwright/test';

test.describe('Authentication & Session Persistence', () => {
	test('should reject invalid login', async ({ page }) => {
		await page.goto('/login');
		await page.fill('input[name="email"]', 'wrong@example.com');
		await page.fill('input[name="password"]', 'badpass');
		await page.click('button[type="submit"]');

		const errorMsg = page.locator('text=Invalid email or password');
		await expect(errorMsg).toBeVisible();
	});

	test('should allow valid login and persist session', async ({ page }) => {
		await page.goto('/login');
		await page.fill('input[name="email"]', 'admin@example.com');
		await page.fill('input[name="password"]', 'password123');
		await page.click('button[type="submit"]');

		await expect(page).toHaveURL(/\/chat\/[^/?]+/);
		await expect(page.getByRole('heading', { name: 'Inbox' })).toBeVisible();
		await expect(page.getByRole('button', { name: /Admin Encanto/ })).toBeVisible();

		await page.reload();
		await expect(page.getByRole('heading', { name: 'Inbox' })).toBeVisible();

		await page.getByRole('button', { name: /Admin Encanto/ }).click();
		await page.getByRole('button', { name: 'Sign out' }).click();
		await expect(page).toHaveURL(/\/login/);

		await page.goto('/chat').catch(() => {});
		await expect(page).toHaveURL(/\/login/);
	});
});
