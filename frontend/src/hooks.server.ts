import { redirect, type Handle } from '@sveltejs/kit';

import type { CurrentUserContext } from '$lib/types';
import { applySetCookies, backendFetch } from '$lib/server/backend';

async function fetchCurrentUser(event: Parameters<Handle>[0]['event']) {
	const meResponse = await backendFetch(event, '/api/me');
	if (meResponse.ok) {
		const payload = (await meResponse.json()) as { data: { user: CurrentUserContext } };
		return payload.data.user;
	}

	const refreshCookie = event.cookies.get('encanto_refresh');
	if (!refreshCookie) {
		return null;
	}

	const refreshResponse = await backendFetch(event, '/api/auth/refresh', { method: 'POST' });
	if (!refreshResponse.ok) {
		return null;
	}
	applySetCookies(event, refreshResponse);

	const retried = await backendFetch(event, '/api/me');
	if (!retried.ok) {
		return null;
	}
	const payload = (await retried.json()) as { data: { user: CurrentUserContext } };
	return payload.data.user;
}

export const handle: Handle = async ({ event, resolve }) => {
	event.locals.user = await fetchCurrentUser(event);

	const pathname = event.url.pathname;
	const isProtected = pathname.startsWith('/chat') || pathname.startsWith('/settings') || pathname.startsWith('/profile');

	if (isProtected && !event.locals.user) {
		throw redirect(303, '/login');
	}

	return resolve(event);
};

