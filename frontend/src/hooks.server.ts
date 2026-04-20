import type { Handle } from '@sveltejs/kit';
import { redirect } from '@sveltejs/kit';
import { resolveApiBaseForHost } from '$lib/api-base';

export const handle: Handle = async ({ event, resolve }) => {
	const API_BASE = resolveApiBaseForHost(event.url.hostname, event.url.protocol);
	const sessionCookie = event.cookies.get('session_token');
	const orgContext = event.cookies.get('org_context');
	const pathname = event.url.pathname;
	const protectedPrefixes = ['/chat', '/settings', '/profile', '/analytics', '/campaigns', '/license-cleanup'];

	if (sessionCookie) {
		try {
			const cookieHeader = [`session_token=${sessionCookie}`, orgContext ? `org_context=${orgContext}` : '']
				.filter(Boolean)
				.join('; ');

			const res = await fetch(`${API_BASE}/api/me`, {
				headers: {
					Cookie: cookieHeader
				}
			});
			if (res.ok) {
				const data = await res.json();
				event.locals.user = data.user;
			} else {
				event.locals.user = null;
			}
		} catch {
			event.locals.user = null;
		}
	} else {
		event.locals.user = null;
	}

	if (protectedPrefixes.some((prefix) => pathname.startsWith(prefix))) {
		if (!event.locals.user) {
			throw redirect(303, '/login');
		}
	}

	if (event.locals.user && sessionCookie) {
		try {
			const cookieHeader = [`session_token=${sessionCookie}`, orgContext ? `org_context=${orgContext}` : '']
				.filter(Boolean)
				.join('; ');
			const res = await fetch(`${API_BASE}/api/license/bootstrap`, {
				headers: {
					Cookie: cookieHeader
				}
			});
			if (res.ok) {
				const bootstrap = await res.json();
				const allowCleanupScreens =
					pathname.startsWith('/settings/license') || pathname.startsWith('/license-cleanup');
				if (bootstrap.restricted_cleanup && !allowCleanupScreens) {
					throw redirect(303, bootstrap.cleanup_url || '/license-cleanup');
				}
			}
		} catch (error) {
			if (
				error &&
				typeof error === 'object' &&
				'status' in error &&
				'location' in error
			) {
				throw error;
			}
		}
	}

	return resolve(event);
};
