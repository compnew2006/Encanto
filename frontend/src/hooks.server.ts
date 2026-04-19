import type { Handle } from '@sveltejs/kit';
import { redirect } from '@sveltejs/kit';
import { env } from '$env/dynamic/public';

const API_BASE = (env.PUBLIC_API_BASE || 'http://127.0.0.1:8080').replace(/\/$/, '');

export const handle: Handle = async ({ event, resolve }) => {
	const sessionCookie = event.cookies.get('session_token');
	const orgContext = event.cookies.get('org_context');

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

	const pathname = event.url.pathname;
	if (pathname.startsWith('/chat') || pathname.startsWith('/settings') || pathname.startsWith('/profile')) {
		if (!event.locals.user) {
			throw redirect(303, '/login');
		}
	}

	return resolve(event);
};
