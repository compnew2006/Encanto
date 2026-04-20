package core

import (
	"fmt"
	"sort"

	"encanto/shared"
)

type PermissionCatalog struct {
	entries map[PermissionKey]shared.PermissionCatalogEntry
	ordered []shared.PermissionCatalogEntry
}

func NewPermissionCatalog() (PermissionCatalog, error) {
	raw, err := shared.LoadPermissionCatalog()
	if err != nil {
		return PermissionCatalog{}, err
	}

	entries := make(map[PermissionKey]shared.PermissionCatalogEntry, len(raw.Permissions))
	for _, permission := range raw.Permissions {
		key := PermissionKey(permission.Key)
		if _, exists := entries[key]; exists {
			return PermissionCatalog{}, fmt.Errorf("duplicate permission key %q", permission.Key)
		}
		entries[key] = permission
	}

	ordered := append([]shared.PermissionCatalogEntry(nil), raw.Permissions...)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].Resource == ordered[j].Resource {
			return ordered[i].Key < ordered[j].Key
		}
		return ordered[i].Resource < ordered[j].Resource
	})

	return PermissionCatalog{entries: entries, ordered: ordered}, nil
}

func (c PermissionCatalog) All() []shared.PermissionCatalogEntry {
	return append([]shared.PermissionCatalogEntry(nil), c.ordered...)
}

func (c PermissionCatalog) Has(key PermissionKey) bool {
	_, exists := c.entries[key]
	return exists
}

func (c PermissionCatalog) Keys() []PermissionKey {
	keys := make([]PermissionKey, 0, len(c.ordered))
	for _, entry := range c.ordered {
		keys = append(keys, PermissionKey(entry.Key))
	}
	return keys
}
