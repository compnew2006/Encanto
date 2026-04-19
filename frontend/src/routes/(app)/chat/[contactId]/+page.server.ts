import { error } from '@sveltejs/kit';

import type { PageServerLoad } from './$types';
import type { ChatDetail, ConversationMessage, ConversationNote } from '$lib/types';
import { backendFetch } from '$lib/server/backend';

export const load: PageServerLoad = async (event) => {
	const contactID = event.params.contactId;
	const chatResponse = await backendFetch(event, `/api/chats/${contactID}`);
	if (chatResponse.status === 404) {
		throw error(404, 'This chat is outside your current visibility scope.');
	}
	if (!chatResponse.ok) {
		throw error(chatResponse.status, 'Unable to load this chat.');
	}

	const messagesResponse = await backendFetch(event, `/api/contacts/${contactID}/messages`);
	const notesResponse = await backendFetch(event, `/api/contacts/${contactID}/notes`);

	const chatPayload = (await chatResponse.json()) as { data: { chat: ChatDetail } };
	const messagesPayload = (await messagesResponse.json()) as { data: { messages: ConversationMessage[] } };
	const notesPayload = (await notesResponse.json()) as { data: { notes: ConversationNote[] } };

	return {
		detail: {
			...chatPayload.data.chat,
			messages: messagesPayload.data.messages,
			notes: notesPayload.data.notes
		}
	};
};

