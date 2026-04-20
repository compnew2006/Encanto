package api

import "strings"

// normalizePhoneNumber strips spaces, dashes, and parens for consistent comparison/storage.
func normalizePhoneNumber(phone string) string {
	phone = extractJIDIdentifier(phone)
	phone = strings.TrimSpace(phone)
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	return phone
}

func extractJIDIdentifier(jid string) string {
	jid = strings.TrimSpace(jid)
	if atIdx := strings.Index(jid, "@"); atIdx != -1 {
		jid = jid[:atIdx]
	}
	if colonIdx := strings.Index(jid, ":"); colonIdx != -1 {
		jid = jid[:colonIdx]
	}
	return jid
}

func isGroupJID(jid string) bool {
	return strings.HasSuffix(strings.TrimSpace(jid), "@g.us")
}

func displayPhoneNumber(phone string) string {
	normalized := normalizePhoneNumber(phone)
	if normalized == "" {
		return ""
	}
	if strings.HasPrefix(normalized, "+") {
		return normalized
	}
	if len(normalized) >= 10 && len(normalized) <= 15 {
		return "+" + normalized
	}
	return normalized
}
