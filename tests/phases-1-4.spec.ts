import { expect, test, type Page } from '@playwright/test';

const chatAliceID = 'f1111111-1111-1111-1111-111111111111';
const chatBobID = 'f2222222-2222-2222-2222-222222222222';
const chatCoraID = 'f3333333-3333-3333-3333-333333333333';
const orgBetaID = '22222222-2222-2222-2222-222222222222';

test.describe.configure({ mode: 'serial' });

async function signIn(page: Page, email: string, password = 'password123') {
	await page.goto('/login');
	await page.getByLabel('Email').fill(email);
	await page.getByLabel('Password').fill(password);
	await Promise.all([
		page.waitForURL('**/chat'),
		page.getByRole('button', { name: 'Sign in' }).click()
	]);
}

async function signOut(page: Page) {
	await page.context().request.post('http://127.0.0.1:5173/logout');
	await page.goto('/chat');
	await expect(page).toHaveURL(/\/login$/);
}

test('redirects anonymous users to the login screen', async ({ page }) => {
	await page.goto('/chat');

	await expect(page).toHaveURL(/\/login$/);
	await expect(page.getByText('Open your current organization context')).toBeVisible();
});

test('switches organizations and hides settings outside the admin context', async ({ page }) => {
	await signIn(page, 'internal@example.com');

	await expect(page.locator('aside').getByText('Encanto Alpha', { exact: true }).first()).toBeVisible();
	await expect(page.getByRole('link', { name: 'Settings' })).toBeVisible();

	await page.locator('aside select').first().selectOption(orgBetaID);
	await expect(page).toHaveURL(/\/chat$/);
	await expect(page.locator('aside').getByText('Encanto Beta', { exact: true }).first()).toBeVisible();
	await expect(page.getByRole('link', { name: /Settings|الإعدادات/ })).toHaveCount(0);
});

test('keeps read-only chat composer disabled and hides pending chats', async ({ page }) => {
	await signIn(page, 'readonly@example.com');

	await expect(page.getByText('Bob Pending')).toHaveCount(0);
	await page.goto(`/chat/${chatAliceID}`);

	await expect(page.getByRole('button', { name: 'Send' })).toBeDisabled();
	await expect(page.getByText('You do not have permission to send messages in this workspace.')).toBeVisible();

	await signOut(page);
});

test('enforces scoped chat visibility and blocks direct access outside scope', async ({ page }) => {
	await signIn(page, 'scoped@example.com');

	await expect(page.getByText('Alice Scope')).toBeVisible();
	await expect(page.getByText('Cora Allowed')).toBeVisible();
	await expect(page.getByText('Bob Pending')).toHaveCount(0);

	const response = await page.goto(`/chat/${chatBobID}`);
	expect(response?.status()).toBe(404);
	await expect(page.getByText('This chat is outside your current visibility scope.')).toBeVisible();
});

test('applies user overrides after a fresh login', async ({ page }) => {
	await signIn(page, 'internal@example.com');
	await page.goto('/settings/users');

	const readOnlyCard = page.locator('section').filter({ has: page.getByText('Read Only Agent') }).first();
	await readOnlyCard.locator('label').filter({ hasText: 'messages.send' }).locator('select').selectOption('allow');
	await readOnlyCard.getByRole('button', { name: 'Save user overrides' }).click();
	await page.waitForLoadState('networkidle');

	await signOut(page);

	await signIn(page, 'readonly@example.com');
	await page.goto(`/chat/${chatAliceID}`);
	await page.reload();

	await expect(page.getByRole('button', { name: 'Send' })).toBeEnabled();
	await expect(page.getByText('You do not have permission to send messages in this workspace.')).toHaveCount(0);

	await signOut(page);

	await signIn(page, 'internal@example.com');
	await page.goto('/settings/users');
	const resetCard = page.locator('section').filter({ has: page.getByText('Read Only Agent') }).first();
	await resetCard.locator('label').filter({ hasText: 'messages.send' }).locator('select').selectOption('inherit');
	await resetCard.getByRole('button', { name: 'Save user overrides' }).click();
	await page.waitForLoadState('networkidle');
	await signOut(page);
});
