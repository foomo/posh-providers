package npm

type Workspaces struct {
	Packages []string           `json:"packages"`
	Catalog  Catalog            `json:"catalog"`
	Catalogs map[string]Catalog `json:"catalogs"`
}
