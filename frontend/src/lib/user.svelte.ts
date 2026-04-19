import { setContext, getContext } from 'svelte';
import type { User } from '../app';

const USER_KEY = Symbol('USER');

export class UserState {
	id = $state('');
	email = $state('');
	name = $state('');
	avatar = $state('');
	status = $state<'online' | 'offline' | 'busy'>('offline');
	role = $state<'admin' | 'agent'>('agent');
	
	settings = $state({
		theme: 'light',
		language: 'ar',
		sidebar_pinned: true
	});

	constructor(user: User) {
		this.update(user);
	}

	update(user: User) {
		this.id = user.id;
		this.email = user.email;
		this.name = user.name;
		this.avatar = user.avatar;
		this.status = user.status;
		this.role = user.role;
		
		this.settings.theme = user.settings.theme;
		this.settings.language = user.settings.language;
		this.settings.sidebar_pinned = user.settings.sidebar_pinned;
	}
}

export function setUserContext(user: User) {
	const userState = new UserState(user);
	setContext(USER_KEY, userState);
	return userState;
}

export function getUserContext(): UserState {
	return getContext<UserState>(USER_KEY);
}
