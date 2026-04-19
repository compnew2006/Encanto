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
		// 1. Login
		await page.goto('/login');
		await page.fill('input[name="email"]', 'admin@example.com');
		await page.fill('input[name="password"]', 'password123');
		await page.click('button[type="submit"]');

		// 2. Verify redirect and secure view
		await expect(page).toHaveURL(/\/chat/);
		await expect(page.locator('text=Welcome to your Inbox')).toBeVisible();
		await expect(page.locator('text=admin@example.com')).toBeVisible();

		// 3. Verify session persistence on reload
		await page.reload();
		await expect(page.locator('text=Welcome to your Inbox')).toBeVisible();

		// 4. Test Logout
		await page.click('text=Logout');
		await expect(page).toHaveURL(/\/login/);

		// 5. Test protected route redirect
		await page.goto('/chat');
		await expect(page).toHaveURL(/\/login/);
	});
});
