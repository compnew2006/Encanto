import { getContext, setContext } from 'svelte';

import type { CurrentUserContext } from '$lib/types';

const USER_CONTEXT = Symbol('user-context');

export class UserState {
	value = $state<CurrentUserContext | null>(null);

	constructor(user: CurrentUserContext | null) {
		this.value = user;
	}

	update(next: CurrentUserContext | null) {
		this.value = next;
	}
}

export function setUserContext(user: CurrentUserContext | null) {
	const state = new UserState(user);
	setContext(USER_CONTEXT, state);
	return state;
}

export function getUserContext() {
	return getContext<UserState>(USER_CONTEXT);
}

