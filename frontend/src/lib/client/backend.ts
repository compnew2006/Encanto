type Envelope<T> = {
	data?: T;
	error?: {
		code: string;
		message: string;
		denialReason?: string;
	};
};

const BACKEND_ORIGIN = 'http://127.0.0.1:58080';

export async function clientJSON<T>(path: string, init?: RequestInit): Promise<T> {
	const response = await fetch(`${BACKEND_ORIGIN}${path}`, {
		...init,
		credentials: 'include',
		headers: {
			'content-type': 'application/json',
			...(init?.headers ?? {})
		}
	});

	const payload = (await response.json()) as Envelope<T>;
	if (!response.ok || payload.error) {
		throw new Error(payload.error?.denialReason ?? payload.error?.message ?? 'Request failed.');
	}
	return payload.data as T;
}
