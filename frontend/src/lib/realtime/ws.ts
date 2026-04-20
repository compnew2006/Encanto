import { apiFetch } from '$lib/api';
import { resolveWsBase } from '$lib/api-base';

type RealtimeListener = (message: { type: string; payload: any }) => void;

export async function connectRealtime(listener: RealtimeListener) {
	const tokenData = await apiFetch<{ token: string }>('/api/auth/ws-token');
	const socket = new WebSocket(`${resolveWsBase()}/ws`, 'whm.v1');

	const heartbeat = () => {
		if (socket.readyState === WebSocket.OPEN) {
			socket.send(JSON.stringify({ type: 'ping', payload: {} }));
		}
	};

	const interval = window.setInterval(heartbeat, 30000);

	socket.addEventListener('open', () => {
		socket.send(JSON.stringify({ type: 'auth', payload: { token: tokenData.token } }));
	});

	socket.addEventListener('message', (event) => {
		const message = JSON.parse(event.data);
		if (message.type === 'auth_ok' || message.type === 'pong') {
			return;
		}
		listener(message);
	});

	return () => {
		window.clearInterval(interval);
		socket.close();
	};
}
