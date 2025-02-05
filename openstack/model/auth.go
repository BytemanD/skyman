package model

type Domain struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type User struct {
	Id          string   `json:"id,omitempty"`
	Name        string   `json:"name"`
	Password    string   `json:"password"`
	Project     string   `json:"project,omitempty"`
	Description string   `json:"description,omitempty"`
	Email       string   `json:"email,omitempty"`
	Enabled     bool     `json:"enabled,omitempty"`
	Domain      Domain   `json:"domain"`
	DomainId    string   `json:"domain_id,omitempty"`
	IsDomain    bool     `json:"is_domain,omitempty"`
	ParentId    bool     `json:"parent_id,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type Password struct {
	User User `json:"user"`
}

type Identity struct {
	Methods  []string `json:"methods,omitempty"`
	Password Password `json:"password,omitempty"`
}

type Project struct {
	Id          string   `json:"id,omitempty"`
	Name        string   `json:"name,omitempty"`
	Domain      Domain   `json:"domain,omitempty"`
	Description string   `json:"description,omitempty"`
	Enabled     bool     `json:"enabled,omitempty"`
	DomainId    string   `json:"domain_id,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	IsDomain    bool     `json:"is_domain,omitempty"`
	ParentId    string   `json:"parent_id,omitempty"`
}
type Scope struct {
	Project Project `json:"project,omitempty"`
}
type Endpoint struct {
	Id        string `json:"id"`
	Region    string `json:"region"`
	Url       string `json:"url"`
	Interface string `json:"interface"`
	RegionId  string `json:"region_id"`
	ServiceId string `json:"service_id"`
}

type Catalog struct {
	Type      string     `json:"type"`
	Name      string     `json:"name"`
	Id        string     `json:"id"`
	Endpoints []Endpoint `json:"endpoints"`
}
type Role struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
type Token struct {
	IsDomain  bool      `json:"is_domain"`
	Methods   []string  `json:"methods"`
	ExpiresAt string    `json:"expires_at"`
	Catalogs  []Catalog `json:"catalog"`
	Roles     []Role    `json:"roles"`
	Project   Project
	User      User
	TokenId   string
}

type Auth struct {
	Identity Identity `json:"identity,omitempty"`
	Scope    Scope    `json:"scope,omitempty"`
}

type AuthBody struct {
	Auth Auth `json:"auth"`
}

type RespToken struct {
	Token Token `json:"token"`
}
