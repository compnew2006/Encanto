import type { PageServerLoad } from './$types';

import { backendFetch } from '$lib/server/backend';
import type { PermissionCatalogEntry, RoleDefinition } from '$lib/types';

export const load: PageServerLoad = async (event) => {
	const [rolesResponse, permissionsResponse] = await Promise.all([
		backendFetch(event, '/api/roles'),
		backendFetch(event, '/api/permissions')
	]);

	const rolesPayload = (await rolesResponse.json()) as { data: { roles: RoleDefinition[] } };
	const permissionsPayload = (await permissionsResponse.json()) as { data: { permissions: PermissionCatalogEntry[] } };

	return {
		roles: rolesPayload.data.roles,
		permissions: permissionsPayload.data.permissions
	};
};

