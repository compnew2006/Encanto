import { browser } from '$app/environment';
import { env } from '$env/dynamic/public';

const FALLBACK_API_BASE = 'http://127.0.0.1:8080';

function trimTrailingSlash(value: string) {
	return value.replace(/\/$/, '');
}

function inferApiBase(hostname: string, protocol = 'http:') {
	const apiProtocol = protocol === 'https:' ? 'https:' : 'http:';
	return `${apiProtocol}//${hostname}:8080`;
}

export function resolveApiBaseForHost(hostname: string, protocol = 'http:') {
	return trimTrailingSlash(env.PUBLIC_API_BASE || inferApiBase(hostname, protocol));
}

export function resolveApiBase() {
	if (env.PUBLIC_API_BASE) {
		return trimTrailingSlash(env.PUBLIC_API_BASE);
	}

	if (browser) {
		return trimTrailingSlash(inferApiBase(window.location.hostname, window.location.protocol));
	}

	return FALLBACK_API_BASE;
}

export function resolveWsBase() {
	const wsUrl = new URL(resolveApiBase());
	wsUrl.protocol = wsUrl.protocol === 'https:' ? 'wss:' : 'ws:';
	return trimTrailingSlash(wsUrl.toString());
}
