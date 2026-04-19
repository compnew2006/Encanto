export type PermissionKey = string;

export type VisibilityScopeMode =
	| 'all_contacts'
	| 'instances_only'
	| 'allowed_numbers_only'
	| 'instances_plus_allowed_numbers';

export type PermissionCatalogEntry = {
	key: PermissionKey;
	resource: string;
	action: string;
	label: string;
	description: string;
};

export type UserSettings = {
	theme: 'light' | 'dark' | 'system';
	language: 'en' | 'ar';
	sidebarPinned: boolean;
};

export type UserOrganization = {
	id: string;
	name: string;
	roleId: string;
	roleName: string;
	isDefault: boolean;
};

export type VisibilityScope = {
	mode: VisibilityScopeMode;
	allowedInstanceIds: string[];
	allowedPhoneNumbers: string[];
	canViewUnmaskedPhone: boolean;
};

export type EffectiveAccess = {
	permissionKeys: PermissionKey[];
	visibility: VisibilityScope;
};

export type CurrentUserContext = {
	id: string;
	email: string;
	fullName: string;
	avatarUrl: string;
	availabilityStatus: 'available' | 'unavailable' | 'busy';
	settings: UserSettings;
	organizations: UserOrganization[];
	currentOrganization: UserOrganization;
	effectiveAccess: EffectiveAccess;
};

export type ChatListItem = {
	id: string;
	name: string;
	phoneNumber: string;
	visiblePhone: string;
	status: 'assigned' | 'pending' | 'closed';
	lastMessagePreview: string;
	lastMessageAt: string | null;
	instanceName: string;
	isHidden: boolean;
	isPinned: boolean;
};

export type ConversationMessage = {
	id: string;
	contactId: string;
	direction: 'inbound' | 'outbound';
	type: string;
	body: string;
	status: string;
	createdAt: string;
	sentByUserId?: string | null;
};

export type ConversationNote = {
	id: string;
	contactId: string;
	authorUserId: string;
	authorName: string;
	body: string;
	createdAt: string;
};

export type ComposerState = {
	allowed: boolean;
	disabled: boolean;
	denialReason: string;
};

export type ChatDetail = {
	chat: ChatListItem;
	messages: ConversationMessage[];
	notes: ConversationNote[];
	composer: ComposerState;
};

export type RoleDefinition = {
	id: string;
	name: string;
	description: string;
	isSystem: boolean;
	isDefault: boolean;
	permissionKeys: PermissionKey[];
	defaultVisibility: VisibilityScope;
};

export type UserSummary = {
	id: string;
	email: string;
	fullName: string;
	avatarUrl: string;
	isActive: boolean;
	availabilityStatus: 'available' | 'unavailable' | 'busy';
	roleId: string;
	roleName: string;
	settings: Partial<UserSettings>;
};

export type SendRestrictions = {
	allowPermissionKeys: PermissionKey[];
	denyPermissionKeys: PermissionKey[];
};

export type ContactVisibilityRule = {
	scopeMode: VisibilityScopeMode;
	allowedInstanceIds: string[];
	allowedPhoneNumbers: string[];
	inheritRoleScope: boolean;
	canViewUnmaskedPhone: boolean;
};

