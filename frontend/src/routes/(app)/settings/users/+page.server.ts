import type { PageServerLoad } from './$types';

import { backendFetch } from '$lib/server/backend';
import type { ContactVisibilityRule, SendRestrictions, UserSummary } from '$lib/types';

export const load: PageServerLoad = async (event) => {
	const usersResponse = await backendFetch(event, '/api/users');
	const usersPayload = (await usersResponse.json()) as { data: { users: UserSummary[] } };
	const users = usersPayload.data.users;

	const hydrated = await Promise.all(
		users.map(async (user) => {
			const [sendResponse, visibilityResponse] = await Promise.all([
				backendFetch(event, `/api/users/${user.id}/send-restrictions`),
				backendFetch(event, `/api/users/${user.id}/contact-visibility`)
			]);
			const sendPayload = (await sendResponse.json()) as { data: SendRestrictions };
			const visibilityPayload = (await visibilityResponse.json()) as { data: { rule: ContactVisibilityRule } };

			return {
				...user,
				sendRestrictions: sendPayload.data,
				visibilityRule: visibilityPayload.data.rule
			};
		})
	);

	return {
		users: hydrated
	};
};

