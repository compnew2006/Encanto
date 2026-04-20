import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import PermissionButton from './PermissionButton.svelte';

describe('PermissionButton', () => {
	it('disables the button and exposes the denial reason', () => {
		render(PermissionButton, {
			allowed: false,
			label: 'Send',
			reason: 'You cannot reply to pending chats until they are claimed.'
		});

		const button = screen.getByRole('button', { name: 'Send' });
		expect((button as HTMLButtonElement).disabled).toBe(true);
		expect(button.getAttribute('title')).toBe('You cannot reply to pending chats until they are claimed.');
	});
});
