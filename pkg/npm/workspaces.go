package npm

// Workspaces represents the "workspaces" section of a package.json, including
// pnpm's catalog extensions.
type Workspaces struct {
	Packages []string           `json:"packages"`
	Catalog  Catalog            `json:"catalog"`
	Catalogs map[string]Catalog `json:"catalogs"`
}
