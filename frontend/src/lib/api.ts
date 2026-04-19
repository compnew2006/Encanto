import { resolveApiBase } from '$lib/api-base';

export const API_BASE = resolveApiBase();

type ApiInit = Omit<RequestInit, 'body'> & {
	body?: BodyInit | object | null;
};

export async function apiFetch<T>(path: string, init: ApiInit = {}): Promise<T> {
	const headers = new Headers(init.headers);
	const hasJSONBody =
		init.body !== undefined &&
		init.body !== null &&
		typeof init.body === 'object' &&
		!(init.body instanceof FormData) &&
		!(init.body instanceof URLSearchParams) &&
		!(init.body instanceof Blob) &&
		!(init.body instanceof ArrayBuffer);

	if (hasJSONBody && !headers.has('Content-Type')) {
		headers.set('Content-Type', 'application/json');
	}

	const response = await fetch(`${API_BASE}${path}`, {
		...init,
		credentials: 'include',
		headers,
		body: hasJSONBody ? JSON.stringify(init.body) : (init.body as BodyInit | null | undefined)
	});

	const text = await response.text();
	const data = text ? JSON.parse(text) : null;

	if (!response.ok) {
		const message = data?.error ?? response.statusText ?? 'Request failed';
		throw new Error(message);
	}

	return data as T;
}

export type WorkspaceUser = {
	id: string;
	name: string;
	email: string;
	role: string;
	status: string;
	avatar: string;
	is_active: boolean;
};

export type ChatContact = {
	id: string;
	organization_id: string;
	name: string;
	phone_number: string;
	phone_display: string;
	avatar: string;
	status: 'assigned' | 'pending' | 'closed' | string;
	assigned_user_id: string;
	assigned_user_name: string;
	instance_id: string;
	instance_name: string;
	instance_source_label: string;
	last_message_preview: string;
	last_message_at: string;
	last_inbound_at: string;
	closed_at?: string | null;
	is_public: boolean;
	is_read: boolean;
	is_pinned: boolean;
	is_hidden: boolean;
	unread_count: number;
	tags: string[];
	metadata: Record<string, string>;
};

export type ChatMessage = {
	id: string;
	contact_id: string;
	direction: 'inbound' | 'outbound' | string;
	type: 'text' | 'media' | string;
	body: string;
	status: string;
	file_name?: string;
	file_size_label?: string;
	media_url?: string;
	failure_reason?: string;
	retry_count: number;
	typed_for_ms: number;
	created_at: string;
	revoked_at?: string | null;
	can_retry: boolean;
	can_revoke: boolean;
	reaction?: string;
};

export type ConversationNote = {
	id: string;
	user_id: string;
	user_name: string;
	body: string;
	created_at: string;
};

export type Collaborator = {
	id: string;
	user_id: string;
	user_name: string;
	status: string;
	invited_at: string;
};

export type TimelineEvent = {
	id: string;
	event_type: string;
	actor_user_id: string;
	actor_name: string;
	summary: string;
	occurred_at: string;
	metadata: Record<string, string>;
};

export type QuickReply = {
	id: string;
	shortcut: string;
	title: string;
	body: string;
};

export type UserNotification = {
	id: string;
	title: string;
	body: string;
	severity: string;
	related_contact_id?: string;
	related_path?: string;
	is_read: boolean;
	created_at: string;
};

export type StatusPost = {
	id: string;
	contact_id: string;
	contact_name: string;
	instance_id: string;
	instance_name: string;
	body: string;
	kind: string;
	created_at: string;
};

export type GeneralSettings = {
	organization_name: string;
	slug: string;
	timezone: string;
	date_format: string;
	locale: string;
	mask_phone_numbers: boolean;
	tenant_status: string;
	active_members: number;
	max_members: number;
	used_instances: number;
	max_instances: number;
	storage_used_label: string;
	storage_limit_label: string;
};

export type AppearanceSettings = {
	color_mode: string;
	theme_preset: string;
};

export type ChatSettings = {
	media_grouping_window_minutes: number;
	sidebar_contact_view: string;
	sidebar_hover_expand: boolean;
	pin_sidebar: boolean;
	chat_background: string;
	show_print_buttons: boolean;
	show_download_buttons: boolean;
};

export type NotificationSettings = {
	email_notifications: boolean;
	new_message_alerts: boolean;
	notification_sound: string;
	campaign_updates: boolean;
};

export type CleanupSettings = {
	retention_days: number;
	run_hour: number;
	timezone: string;
	last_run_at?: string | null;
	last_job_status: string;
};

export type SettingsSummary = {
	general: GeneralSettings;
	appearance: AppearanceSettings;
	chat: ChatSettings;
	notifications: NotificationSettings;
	cleanup: CleanupSettings;
};

export type WhatsAppInstance = {
	id: string;
	organization_id: string;
	name: string;
	phone_number: string;
	jid: string;
	status: string;
	pairing_state: string;
	qr_code?: string;
	slot_blocked: boolean;
	settings: {
		auto_sync_history: boolean;
		auto_download_incoming_media: boolean;
		source_tag_label: string;
		source_tag_display_mode: string;
		source_tag_color: string;
	};
	health: {
		status: string;
		uptime_label: string;
		queue_depth: number;
		sent_today: number;
		received_today: number;
		failed_today: number;
		error_rate: string;
		observed_at: string;
	};
	call_policy: {
		enabled: boolean;
		reject_individual_calls: boolean;
		reject_group_calls: boolean;
		reply_mode: string;
		schedule_mode: string;
		emergency_bypass: string[];
		reply_message: string;
	};
	auto_campaign: {
		enabled: boolean;
		campaign_name_prefix: string;
		schedule_every_days: number;
		delay_from_minutes: number;
		delay_to_minutes: number;
		campaign_status: string;
		message_body: string;
	};
	rating_settings: {
		enabled: boolean;
		follow_up_window_minutes: number;
		template_ar: string;
		template_en: string;
	};
	assignment_reset: {
		enabled: boolean;
		schedule_mode: string;
		timezone: string;
	};
};

export type WorkspaceSnapshot = {
	current_tab: string;
	tab_counts: Record<string, number>;
	filters: Record<string, string>;
	conversations: ChatContact[];
	selected?: {
		contact: ChatContact;
		messages: ChatMessage[];
		notes: ConversationNote[];
		collaborators: Collaborator[];
		events: TimelineEvent[];
	};
	notifications: UserNotification[];
	statuses: StatusPost[];
	quick_replies: QuickReply[];
	instances: WhatsAppInstance[];
	users: WorkspaceUser[];
	settings: SettingsSummary;
};

export type ProfileView = {
	user: App.User;
	settings: SettingsSummary;
};

export type InstanceHealthSummary = {
	id: string;
	name: string;
	status: string;
	uptime_label: string;
	queue_depth: number;
	sent_today: number;
	received_today: number;
	failed_today: number;
	error_rate: string;
	observed_at: string;
};

export function formatDateTime(value?: string | null) {
	if (!value) return 'Never';
	return new Date(value).toLocaleString();
}
