import { expect, test, type Page } from '@playwright/test';

declare const Buffer: {
	from(input: string): Uint8Array;
};

async function login(page: Page) {
	await page.goto('/login');
	await page.fill('input[name="email"]', 'admin@example.com');
	await page.fill('input[name="password"]', 'password123');
	await page.click('button[type="submit"]');
	await expect(page).toHaveURL(/\/chat\/[^/?]+/);
}

test.describe('Operational flows', () => {
	test('updates profile and organization settings', async ({ page }) => {
		await login(page);

		await page.goto('/profile');
		await page.getByLabel('Display Name').fill('Admin Encanto QA');
		await page.getByTestId('save-profile').click();
		await expect(page.getByText('Profile saved.')).toBeVisible();
		await expect(page.getByLabel('Display Name')).toHaveValue('Admin Encanto QA');

		await page.goto('/settings');
		await page.getByLabel('Organization Name').fill('Global Corp Ops');
		await page.getByTestId('save-general-settings').click();
		await expect(page.getByText('General settings saved.')).toBeVisible();

		await page.getByRole('button', { name: 'Appearance' }).click();
		await page.getByLabel('Theme Preset').selectOption('soft-pop');
		await page.getByTestId('save-appearance-settings').click();
		await expect(page.getByText('Appearance settings saved.')).toBeVisible();

		await page.getByRole('button', { name: 'Notifications' }).click();
		await page.getByLabel('Notification Sound').selectOption('digital-ping');
		await page.getByTestId('save-notification-settings').click();
		await expect(page.getByText('Notification settings saved.')).toBeVisible();

		await page.getByTestId('cleanup-retention-days').fill('21');
		await page.getByTestId('cleanup-run-hour').fill('6');
		await page.getByTestId('save-cleanup-settings').click();
		await expect(page.getByText('Cleanup schedule saved.')).toBeVisible();
		await expect(page.getByTestId('cleanup-retention-days')).toHaveValue('21');
		await expect(page.getByTestId('cleanup-run-hour')).toHaveValue('6');

		await page.getByTestId('run-cleanup').click();
		await expect(page.getByText('Cleanup job completed.')).toBeVisible();
	});

	test('handles chat actions and status workflows', async ({ page }) => {
		await login(page);

		await page.goto('/chat/contact-1');
		await expect(page.getByRole('heading', { name: 'Mina Salah' })).toBeVisible();

		await page.getByTestId('composer-textarea').fill('Playwright follow-up message');
		await page.getByTestId('send-message').click();
		await expect(page.getByText('Message queued.')).toBeVisible();
		await expect(page.getByText('Playwright follow-up message').last()).toBeVisible();

		await page.getByPlaceholder('Add an internal note...').fill('Internal note from Playwright');
		await page.getByTestId('add-note').click();
		await expect(page.getByText('Note added.')).toBeVisible();
		await expect(page.getByText('Internal note from Playwright').last()).toBeVisible();

		await page.getByTestId('notifications-toggle').click();
		await expect(page.getByRole('heading', { name: 'Notifications' })).toBeVisible();
		await page.getByRole('button', { name: 'Mark all as read' }).click();
		await expect(page.getByText('Notifications cleared.')).toBeVisible();
		await page.getByRole('button', { name: 'Close notifications panel' }).evaluate((element) => {
			(element as HTMLDivElement).click();
		});
		await expect(page.getByRole('heading', { name: 'Notifications' })).toBeHidden();

		await page.getByTestId('statuses-toggle').click();
		await page
			.getByPlaceholder('Add a status update for the current conversation or instance...')
			.fill('Status from Playwright');
		await page.getByTestId('add-status').click();
		await expect(page.getByText('Status posted.')).toBeVisible();
		await expect(page.getByText('Status from Playwright').last()).toBeVisible();
	});

	test('manages accounts and policies', async ({ page }) => {
		await login(page);

		await page.goto('/settings/instances');
		await expect(page.getByText('Sales WA')).toBeVisible();
		await expect(page.getByText('Care WA')).toBeVisible();

		await page.getByPlaceholder('New account name').fill('QA Account');
		await page.getByPlaceholder('Phone number').fill('+201555555555');
		await page.getByTestId('create-instance').click();
		await expect
			.poll(() =>
				page.locator('input').evaluateAll((inputs) =>
					inputs.some((input) => input instanceof HTMLInputElement && input.value === 'QA Account')
				)
			)
			.toBe(true);

		await page.getByTestId('connect-inst-2').click();
		await expect(page.getByText('Care WA connected.')).toBeVisible();

		await page.getByTestId('save-instance-inst-1').click();
		await expect(page.getByText('Sales WA settings saved.')).toBeVisible();
	});

	test('creates direct chats and sends media attachments', async ({ page }) => {
		await login(page);

		await page.goto('/chat/contact-1');
		await page.getByTestId('open-direct-chat').click();
		await page.getByTestId('direct-chat-phone').fill('+201566677788');
		await page.getByTestId('direct-chat-name').fill('Playwright Direct');
		await page.getByTestId('direct-chat-instance').selectOption('inst-1');
		await page.getByTestId('create-direct-chat').click();

		await expect(page.getByText('Direct chat created.')).toBeVisible();
		await expect(page.getByRole('heading', { name: 'Playwright Direct' })).toBeVisible();
		await expect(page.getByText('+201566677788').first()).toBeVisible();

		await page.getByTestId('composer-media').click();
		await page.setInputFiles('[data-testid="attachment-file-input"]', {
			name: 'quote.txt',
			mimeType: 'text/plain',
			buffer: Buffer.from('draft quote attachment')
		});
		await expect(page.getByTestId('attachment-preview-name')).toHaveText('quote.txt');
		await page.getByTestId('composer-caption').fill('Attached quote draft');
		await page.getByTestId('send-message').click();

		await expect(page.getByText('Message queued.')).toBeVisible();
		await expect(page.getByText('quote.txt').last()).toBeVisible();
		await expect(page.getByText('Attached quote draft').last()).toBeVisible();
	});
});
