import { redirect, type RequestHandler } from '@sveltejs/kit';

import { backendFetch } from '$lib/server/backend';

export const POST: RequestHandler = async (event) => {
	await backendFetch(event, '/api/auth/logout', { method: 'POST' });
	event.cookies.delete('encanto_access', { path: '/' });
	event.cookies.delete('encanto_refresh', { path: '/' });
	throw redirect(303, '/login');
};
