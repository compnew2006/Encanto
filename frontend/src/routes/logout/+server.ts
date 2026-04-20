import { redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { resolveApiBaseForHost } from '$lib/api-base';

export const POST: RequestHandler = async ({ cookies, url }) => {
	// Erase cookie
	cookies.delete('session_token', { path: '/' });
	cookies.delete('org_context', { path: '/' });

	// Optional: Call backend to invalidate token if stateful sessions are used
	try {
		await fetch(`${resolveApiBaseForHost(url.hostname, url.protocol)}/api/auth/logout`, { method: 'POST' });
	} catch {
		// Ignore self error on logout
	}

	throw redirect(303, '/login');
};
