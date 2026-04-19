import type { Handle } from '@sveltejs/kit';
import { redirect } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
	// 1. Get user from the backend
	const sessionCookie = event.cookies.get('session_token');

	if (sessionCookie) {
		try {
			const res = await fetch('http://127.0.0.1:8080/api/me', {
				headers: {
					'Cookie': `session_token=${sessionCookie}`
				}
			});
			if (res.ok) {
				const data = await res.json();
				console.log("Hooks: Fetch OK", data);
				event.locals.user = data.user;
			} else {
				const t = await res.text();
				console.log("Hooks: Fetch Failed", res.status, t);
				event.locals.user = null;
			}
		} catch (error) {
			console.log("Hooks: Fetch Error", error);
			event.locals.user = null;
		}
	} else {
		console.log("Hooks: No session cookie");
		event.locals.user = null;
	}

	// 2. Protect routes
	const pathname = event.url.pathname;
	if (pathname.startsWith('/chat') || pathname.startsWith('/settings')) {
		if (!event.locals.user) {
			throw redirect(303, '/login');
		}
	}

	return resolve(event);
};
