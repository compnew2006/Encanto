package shared

import (
	"embed"
	"encoding/json"
)

//go:embed permission_catalog.json
var permissionCatalogFS embed.FS

type PermissionCatalogEntry struct {
	Key         string `json:"key"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type PermissionCatalog struct {
	Permissions []PermissionCatalogEntry `json:"permissions"`
}

func LoadPermissionCatalog() (PermissionCatalog, error) {
	raw, err := permissionCatalogFS.ReadFile("permission_catalog.json")
	if err != nil {
		return PermissionCatalog{}, err
	}

	var catalog PermissionCatalog
	if err := json.Unmarshal(raw, &catalog); err != nil {
		return PermissionCatalog{}, err
	}

	return catalog, nil
}

