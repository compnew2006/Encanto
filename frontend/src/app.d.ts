// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		interface UserSettings {
			theme: string;
			language: string;
			sidebar_pinned: boolean;
		}

		interface Organization {
			id: string;
			name: string;
			role: string;
		}

		interface User {
			id: string;
			email: string;
			name: string;
			avatar: string;
			status: 'online' | 'offline' | 'busy' | string;
			role: 'admin' | 'agent' | string;
			settings: UserSettings;
			organizations: Organization[];
			current_organization: Organization;
		}

		interface Locals {
			user: User | null;
		}
	}
}

export {};
