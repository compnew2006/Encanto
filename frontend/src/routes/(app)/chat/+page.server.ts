import type { PageServerLoad } from './$types';

import { backendFetch } from '$lib/server/backend';
import type { ChatListItem } from '$lib/types';

export const load: PageServerLoad = async (event) => {
	const search = event.url.searchParams.get('search') ?? '';
	const response = await backendFetch(event, `/api/chats?search=${encodeURIComponent(search)}`);
	const payload = (await response.json()) as { data: { chats: ChatListItem[] } };

	return {
		search,
		chats: payload.data.chats
	};
};

