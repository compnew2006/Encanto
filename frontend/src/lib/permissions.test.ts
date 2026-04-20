import { describe, expect, it } from 'vitest';

import { denialReason, groupPermissions, hasPermission, permissionCatalog } from '$lib/permissions';
import type { CurrentUserContext } from '$lib/types';

const user: CurrentUserContext = {
	id: '1',
	email: 'admin@example.com',
	fullName: 'Admin',
	avatarUrl: '',
	availabilityStatus: 'available',
	settings: { theme: 'light', language: 'en', sidebarPinned: true },
	organizations: [],
	currentOrganization: {
		id: 'org',
		name: 'Org',
		roleId: 'role',
		roleName: 'Admin',
		isDefault: true
	},
	effectiveAccess: {
		permissionKeys: ['chats.view', 'messages.send'],
		visibility: {
			mode: 'all_contacts',
			allowedInstanceIds: [],
			allowedPhoneNumbers: [],
			canViewUnmaskedPhone: true
		}
	}
};

describe('permission helpers', () => {
	it('checks permissions from the effective access payload', () => {
		expect(hasPermission(user, 'messages.send')).toBe(true);
		expect(hasPermission(user, 'roles.manage')).toBe(false);
	});

	it('returns stable denial copy for known restricted actions', () => {
		expect(denialReason(user, 'messages.send')).toContain('send messages');
		expect(denialReason(user, 'chats.unclaimed.send', 'pending')).toContain('pending chats');
	});

	it('groups the permission catalog by resource', () => {
		const grouped = groupPermissions(permissionCatalog);
		expect(grouped.chats.length).toBeGreaterThan(0);
		expect(grouped.contacts.length).toBeGreaterThan(0);
	});
});

