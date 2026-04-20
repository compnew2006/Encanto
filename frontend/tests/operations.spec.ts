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

	test('creates direct chats and sends media attachments', async ({ page, browserName }) => {
		await login(page);

		const token = Date.now().toString().slice(-6);
		const directName = `Playwright Direct ${browserName} ${token}`;
		const directPhone = `+2015${token}`;

		await page.goto('/chat/contact-1');
		await page.getByTestId('open-direct-chat').click();
		await page.getByTestId('direct-chat-phone').fill(directPhone);
		await page.getByTestId('direct-chat-name').fill(directName);
		await page.getByTestId('direct-chat-instance').selectOption('inst-1');
		await page.getByTestId('create-direct-chat').click();

		await expect(page.getByText('Direct chat created.')).toBeVisible();
		await expect(page.getByRole('heading', { name: directName })).toBeVisible();
		await expect(page.getByText(directPhone).first()).toBeVisible();

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

	test('manages contacts and reopens a closed chat from settings', async ({ page, browserName }) => {
		await login(page);

		const token = Date.now().toString().slice(-6);
		const contactName = `Phase11 ${browserName} ${token}`;
		const updatedName = `${contactName} Updated`;
		const importedName = `Imported ${browserName} ${token}`;
		const phone = `+2016${token}`;
		const importPhone = `+2017${token}`;

		await page.goto('/settings/contacts');
		await expect(page.getByRole('heading', { name: 'Contacts Directory' })).toBeVisible();

		await page.getByTestId('contact-name').fill(contactName);
		await page.getByTestId('contact-phone').fill(phone);
		await page.getByTestId('save-contact').click();
		await expect(page.getByText('Contact created.')).toBeVisible();
		await expect(page.getByText(contactName)).toBeVisible();

		const createdCard = page.locator(`[data-contact-name="${contactName}"]`);
		await createdCard.getByRole('button', { name: 'Edit' }).click();
		await page.getByTestId('contact-name').fill(updatedName);
		await page.getByTestId('save-contact').click();
		await expect(page.getByText('Contact updated.')).toBeVisible();
		await expect(page.getByText(updatedName)).toBeVisible();

		await page.getByTestId('contacts-export').click();
		await expect(page.locator('textarea[readonly]').first()).toHaveValue(new RegExp(updatedName));

		await page.getByTestId('contacts-import-csv').fill(
			`name,phone_number,instance_id,tags\n${importedName},${importPhone},inst-1,vip`
		);
		await page.getByTestId('contacts-import').click();
		await expect(page.getByText('Contacts imported.')).toBeVisible();
		await expect(page.getByText(importedName)).toBeVisible();

		await page.locator(`[data-contact-name="${updatedName}"]`).getByRole('button', { name: 'Open Chat' }).click();
		await expect(page).toHaveURL(/\/chat\/[^/?]+/);
		await expect(page.getByRole('heading', { name: updatedName })).toBeVisible();

		await page.getByTestId('close-action').click();
		await expect(page.getByTestId('close-action')).toHaveText('Reopen');

		await page.goto('/settings/closed-chats');
		await expect(page.getByRole('heading', { name: 'Closed Conversations' })).toBeVisible();
		await page.locator(`[data-contact-name="${updatedName}"]`).getByRole('button', { name: 'Reopen' }).click();
		await expect(page).toHaveURL(/\/chat\/[^/?]+/);
		await expect(page.getByRole('heading', { name: updatedName })).toBeVisible();
	});

	test('activates the license flow and exits restricted cleanup mode', async ({ page, browserName }) => {
		await login(page);

		const token = Date.now().toString().slice(-6);
		const cleanupName = `Cleanup ${browserName} ${token}`;
		const cleanupPhone = `+2018${token}`;

		await page.goto('/settings/contacts');
		await page.getByTestId('contact-name').fill(cleanupName);
		await page.getByTestId('contact-phone').fill(cleanupPhone);
		await page.getByTestId('save-contact').click();
		await expect(page.getByText('Contact created.')).toBeVisible();

		await page.goto('/settings/license');
		await expect(page.getByRole('heading', { name: 'License & Limits' })).toBeVisible();
		await page.getByTestId('license-key').fill(`RESTRICT-CLEANUP-${browserName}-${token}`);
		await page.getByTestId('activate-license').click();
		await expect(page).toHaveURL(/\/license-cleanup/);

		await page.goto('/analytics/agents');
		await expect(page).toHaveURL(/\/license-cleanup/);

		page.once('dialog', (dialog) => dialog.accept());
		await page.locator(`[data-contact-name="${cleanupName}"]`).getByRole('button', { name: 'Delete Contact' }).click();
		await expect(page.locator(`[data-contact-name="${cleanupName}"]`)).toHaveCount(0);
		await expect(page.getByText('Cleanup is complete. Normal operations are available again.')).toBeVisible();

		await page.goto('/settings/license');
		await page.getByTestId('license-key').fill(`NORMAL-LICENSE-${browserName}-${token}`);
		await page.getByTestId('activate-license').click();
		await expect(page.getByText('License activated.')).toBeVisible();
		await expect(page.getByTestId('license-status')).toHaveText('active');

		await page.goto('/settings');
		await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible();
	});

	test('renders analytics, campaigns, and audit operations', async ({ page, browserName }) => {
		await login(page);

		const token = Date.now().toString().slice(-6);
		const campaignName = `Campaign ${browserName} ${token}`;
		const campaignBody = `Playwright launch body ${token}`;

		await page.goto('/analytics/agents');
		await expect(page.getByRole('heading', { name: 'Agent Analytics' })).toBeVisible();
		await page.getByTestId('analytics-export').click();
		await expect(page.locator('textarea[readonly]').first()).toHaveValue(/.+/);

		await page.goto('/campaigns');
		await expect(page.getByRole('heading', { name: 'Campaigns' })).toBeVisible();
		await page.getByTestId('campaign-create-new').click();
		await page.getByTestId('campaign-name').fill(campaignName);
		await page.getByTestId('campaign-content').fill(campaignBody);
		await page.getByTestId('campaign-instance-filter').selectOption('inst-1');
		await page.getByTestId('campaign-every-days').fill('5');
		await page.getByTestId('campaign-save').click();
		await expect(page.getByText('Campaign created.')).toBeVisible();
		await expect(page.getByText(campaignName)).toBeVisible();

		await page.getByTestId('campaign-launch').click();
		await expect(page.getByText('Campaign launched.')).toBeVisible();
		await expect(page.getByText('Mina Salah').last()).toBeVisible();

		await page.goto('/settings/audit');
		await expect(page.getByRole('heading', { name: 'Audit & Reliability' })).toBeVisible();
		await expect(page.getByText('campaign_run').first()).toBeVisible();

		const retryButton = page.locator('[data-testid^="retry-delivery-"]').first();
		await expect(retryButton).toBeVisible();
		await retryButton.click();
		await expect(page.getByText('Webhook delivery retried.')).toBeVisible();
		await expect(page.getByText('campaigns.launch').first()).toBeVisible();
	});
});
