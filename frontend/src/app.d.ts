// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
import type { CurrentUserContext } from '$lib/types';

declare global {
	namespace App {
		interface Locals {
			user: CurrentUserContext | null;
		}

		interface PageData {
			user: CurrentUserContext | null;
		}
	}
}

export {};
