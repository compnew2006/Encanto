import { redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ cookies }) => {
	// Erase cookie
	cookies.delete('session_token', { path: '/' });

	// Optional: Call backend to invalidate token if stateful sessions are used
	try {
        fetch('http://127.0.0.1:8080/api/auth/logout', { method: 'POST' });
    } catch (e) {
        // Ignore self error on logout
    }

	throw redirect(303, '/login');
};
