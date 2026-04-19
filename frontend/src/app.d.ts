// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		interface UserSettings {
			theme: string;
			language: string;
			sidebar_pinned: boolean;
		}

		interface User {
			id: string;
			email: string;
			name: string;
			avatar: string;
			status: 'online' | 'offline' | 'busy';
			role: 'admin' | 'agent';
			settings: UserSettings;
		}

		interface Locals {
			user: User | null;
		}
	}
}

export {};
