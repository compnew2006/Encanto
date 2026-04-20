import catalog from '$backendShared/permission_catalog.json';

import type { CurrentUserContext, PermissionCatalogEntry, PermissionKey, VisibilityScopeMode } from '$lib/types';

export const permissionCatalog = catalog.permissions as PermissionCatalogEntry[];

export function hasPermission(user: CurrentUserContext | null | undefined, key: PermissionKey) {
	return user?.effectiveAccess.permissionKeys.includes(key) ?? false;
}

export function canViewSettings(user: CurrentUserContext | null | undefined) {
	return hasPermission(user, 'settings.manage') || hasPermission(user, 'roles.manage') || hasPermission(user, 'users.view');
}

export function denialReason(user: CurrentUserContext | null | undefined, key: PermissionKey, chatStatus?: string) {
	if (!user) {
		return 'You need to sign in before this action is available.';
	}
	if (key === 'messages.send') {
		return 'You do not have permission to send messages in this workspace.';
	}
	if (key === 'chats.unclaimed.send' && chatStatus === 'pending') {
		return 'You cannot reply to pending chats until they are claimed.';
	}
	if (key === 'roles.manage') {
		return 'Only users with role-management access can change role definitions.';
	}
	if (key === 'users.update') {
		return 'Only users with user-management access can change user overrides.';
	}
	return 'Your current permissions do not allow this action.';
}

export function groupPermissions(entries = permissionCatalog) {
	return entries.reduce<Record<string, PermissionCatalogEntry[]>>((grouped, entry) => {
		grouped[entry.resource] = grouped[entry.resource] ?? [];
		grouped[entry.resource].push(entry);
		return grouped;
	}, {});
}

export function scopeLabel(scope: VisibilityScopeMode) {
	switch (scope) {
		case 'all_contacts':
			return 'All contacts';
		case 'instances_only':
			return 'Instance-only';
		case 'allowed_numbers_only':
			return 'Allowed numbers only';
		case 'instances_plus_allowed_numbers':
			return 'Instances + allowed numbers';
	}
}

