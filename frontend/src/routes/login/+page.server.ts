import { fail, redirect, isRedirect } from '@sveltejs/kit';
import { resolveApiBaseForHost } from '$lib/api-base';
import type { Actions } from './$types';

export const actions = {
	default: async ({ request, cookies, url }) => {
		const API_BASE = resolveApiBaseForHost(url.hostname, url.protocol);
		const data = await request.formData();
		const email = data.get('email');
		const password = data.get('password');

		if (!email || !password) {
			return fail(400, { email, missing: true });
		}

		try {
			const res = await fetch(`${API_BASE}/api/auth/login`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ email, password })
			});

			const payload = await res.json().catch(() => null);

			if (res.ok) {
				const rawSetCookies = (res.headers as Headers & { getSetCookie?: () => string[] }).getSetCookie?.() ?? [];
				const setCookieHeader = rawSetCookies.length ? rawSetCookies.join('; ') : (res.headers.get('set-cookie') ?? '');
				const tokenMatch = setCookieHeader.match(/session_token=([^;]+)/);

				if (!tokenMatch) {
					return fail(500, { email, error: 'Missing session token from backend login response.' });
				}

				cookies.set('session_token', tokenMatch[1], {
					path: '/',
					httpOnly: true,
					sameSite: 'lax',
					secure: false,
					maxAge: 60 * 60 * 24
				});

				const activeOrgId = payload?.user?.current_organization?.id;
				if (activeOrgId) {
					cookies.set('org_context', activeOrgId, {
						path: '/',
						httpOnly: true,
						sameSite: 'lax',
						secure: false,
						maxAge: 60 * 60 * 24 * 365
					});
				}

				throw redirect(303, '/chat');
			}

			return fail(401, { email, incorrect: true, message: payload?.error ?? 'Invalid email or password.' });
		} catch (e) {
			if (isRedirect(e)) {
				throw e;
			}
			return fail(500, { email, error: 'Server error' });
		}
	}
} satisfies Actions;
