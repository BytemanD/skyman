package identity

type Token struct {
	IsDomain  bool      `json:"is_domain"`
	Methods   []string  `json:"methods"`
	ExpiresAt string    `json:"expires_at"`
	Name      bool      `json:"name"`
	Catalogs  []Catalog `json:"catalog"`
	TokenId   string
	Project   Project
	User      User
}
type Catalog struct {
	Type      string     `json:"type"`
	Name      string     `json:"name"`
	Id        string     `json:"id"`
	Endpoints []Endpoint `json:"endpoints"`
}

type Endpoint struct {
	Url       string `json:"url"`
	Interface string `json:"interface"`
	Region    string `json:"region"`
	RegionId  string `json:"region_id"`
	Id        string `json:"id"`
}

type RespToken struct {
	Token Token `json:"token"`
}

type OptionCatalog struct {
	Region    string
	Type      string
	Name      string
	Interface string
}

func (token *RespToken) GetCatalogByType(serviceType string) *Catalog {
	for _, catalog := range token.Token.Catalogs {
		if catalog.Type == serviceType {
			return &catalog
		}
	}
	return &Catalog{}
}
func (token *Token) GetEndpoints(option OptionCatalog) []Endpoint {
	endpoints := []Endpoint{}

	for _, catalog := range token.Catalogs {
		if (option.Type != "" && catalog.Type != option.Type) ||
			(option.Name != "" && catalog.Name != option.Name) {
			continue
		}
		for _, endpoint := range catalog.Endpoints {
			if (option.Region != "" && endpoint.Region != option.Region) ||
				(option.Interface != "" && endpoint.Interface != option.Interface) {
				continue
			}
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}
