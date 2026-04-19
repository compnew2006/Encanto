import type { RequestEvent } from '@sveltejs/kit';

const BACKEND_ORIGIN = 'http://127.0.0.1:58080';

type Envelope<T> = {
	data?: T;
	error?: {
		code: string;
		message: string;
		denialReason?: string;
	};
};

export async function backendFetch(event: RequestEvent, path: string, init?: RequestInit) {
	const headers = new Headers(init?.headers);
	const cookieHeader = event.request.headers.get('cookie');
	if (cookieHeader) {
		headers.set('cookie', cookieHeader);
	}
	if (!headers.has('content-type') && init?.body) {
		headers.set('content-type', 'application/json');
	}
	return fetch(`${BACKEND_ORIGIN}${path}`, {
		...init,
		headers
	});
}

export async function backendJSON<T>(event: RequestEvent, path: string, init?: RequestInit): Promise<Envelope<T>> {
	const response = await backendFetch(event, path, init);
	const payload = (await response.json()) as Envelope<T>;
	return payload;
}

export function applySetCookies(event: RequestEvent, response: Response) {
	const rawCookies = response.headers.getSetCookie?.() ?? [];
	for (const rawCookie of rawCookies) {
		const [nameValue, ...parts] = rawCookie.split(';');
		const [name, value] = nameValue.split('=');
		const normalized = name.trim();
		if (!normalized) {
			continue;
		}

		const cookieOptions = {
			path: '/',
			httpOnly: true as const,
			secure: false,
			sameSite: 'lax' as const,
			maxAge: undefined as number | undefined
		};

		for (const part of parts) {
			const [key, raw] = part.trim().split('=');
			if (key.toLowerCase() === 'max-age' && raw) {
				cookieOptions.maxAge = Number(raw);
			}
		}

		if (value === '' || cookieOptions.maxAge === -1) {
			event.cookies.delete(normalized, { path: '/' });
			continue;
		}

		event.cookies.set(normalized, value, cookieOptions);
	}
}
