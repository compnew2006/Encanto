import { fail, redirect, isRedirect } from '@sveltejs/kit';
import type { Actions } from './$types';

export const actions = {
	default: async ({ request, cookies }) => {
		const data = await request.formData();
		const email = data.get('email');
		const password = data.get('password');

		if (!email || !password) {
			return fail(400, { email, missing: true });
		}

		try {
			const res = await fetch('http://127.0.0.1:8080/api/auth/login', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ email, password })
			});

			if (res.ok) {
				const setCookie = res.headers.get('set-cookie');
				console.log("LOGIN RESPONSE SET-COOKIE:", setCookie);
				if (setCookie) {
					const tokenMatch = setCookie.match(/session_token=([^;]+)/);
					if (tokenMatch) {
						console.log("Extracted session_token:", tokenMatch[1].substring(0, 10) + "...");
						cookies.set('session_token', tokenMatch[1], {
							path: '/',
							httpOnly: true,
							sameSite: 'lax',
							secure: false,
							maxAge: 60 * 60 * 24 // 24 hours
						});
					} else {
						console.log("Failed to match session_token in cookie string");
					}
				}
				throw redirect(303, '/chat');
			} else {
				const errorData = await res.json();
				return fail(401, { email, incorrect: true, message: errorData.error });
			}
		} catch (e) {
			if (isRedirect(e)) {
				throw e;
			}
			console.error("Login catch block:", e);
			return fail(500, { email, error: 'Server error' });
		}
	}
} satisfies Actions;
