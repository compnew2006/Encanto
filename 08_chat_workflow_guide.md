# Whatomate Chat Workflow Guide

Live-audited reference for the current chat workspace. This guide reflects what appeared on the real /chat screen on 2026-04-18, not the earlier reduced-scope assumption.

## Observed Live Scope

- Assigned / pending inbox
- Notifications center
- Statuses drawer
- Start new chat
- Assign / notes / contact info
- Tags / collaborators
- Emoji / quick replies / attachment / print
- Retry / revoke / download
- **Meta** still not confirmed
- **AI / SLA / transfer queue** not confirmed

## Observed Live Layout

- **Left rail**: counters, refresh, notifications, statuses, search, add contact, filters, assigned/pending tabs, chat rows.
- **Center**: conversation header, history, composer.
- **Right side**: notes panel and contact info panel can be open together.

## Inbox Rail

- **Search contacts**: text search for inbox contacts.
- **Add Contact**: opens Start New Chat and asks for phone number, optional profile name, and WhatsApp instance.
- **Filter**: filters by instance, chat type (Private chats, Groups, Channels), and tags.
- **Tabs**: Assigned and Pending.
- **Row actions**: every row exposes Hide chat and Delete chat.

## Conversation Header

- **Assign**: user-plus opens Assign Contact with searchable user list.
- **Pin**: pin icon exists in the header action cluster.
- **Responsibility reset**: a user-x action exists and requires a dedicated backend action in the plan.
- **Notes**: opens internal notes panel.
- **Info**: opens Contact Info with tags, collaborators, and general fields.

## Right Panels

- **Notes**: empty-state prompt plus write box for internal notes.
- **Contact Info**: phone, tags, collaborators, avatar sync fields, and extra panel area.
- **Chatbot dependency**: the extra panel area explicitly says "Configure panel display in the chatbot flow settings".

## Message Composer

- **Emoji**: smile icon.
- **Quick replies**: message-square-text action.
- **Attachment**: paperclip.
- **Print**: printer action.
- **Send**: disabled until valid input exists.

## Message History Actions

- **Media**: images and files expose Print and Download.
- **Batches**: grouped files can expose Download All.
- **Revoke**: many outbound messages expose Revoke message.
- **Retry**: failed messages show Retry sending and an error such as instance not connected.
- **Reactions**: reactions are visible in history.

## Overlays Confirmed In Chat

- **Notifications dialog**: unread messages and system notifications with Mark all as read.
- **Statuses drawer**: Add status plus list of status entries by contact and instance.
- **Assign Contact dialog**: searchable assignee list.
- **Invite Collaborator dialog**: searchable user list for shared chat access.