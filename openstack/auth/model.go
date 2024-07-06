package auth

import (
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

type Domain struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type User struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Password    string `json:"password"`
	Project     string `json:"project,omitempty"`
	Description string `json:"description,omitempty"`
	Email       string `json:"email,omitempty"`
	Enabled     bool   `json:"enabled,omitempty"`
	Domain      Domain `json:"domain"`
	DomainId    string `json:"domain_id,omitempty"`
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

type Token struct {
	IsDomain  bool      `json:"is_domain"`
	Methods   []string  `json:"methods"`
	ExpiresAt string    `json:"expires_at"`
	Name      bool      `json:"name"`
	Catalogs  []Catalog `json:"catalog"`
	Project   Project
	User      User
}

type TokenCache struct {
	token     Token
	TokenId   string
	expiredAt time.Time
}

func (tc *TokenCache) IsTokenExpired() bool {
	if tc.TokenId == "" {
		return true
	}
	if tc.expiredAt.Before(time.Now()) {
		logging.Debug("token is exipred")
		return true
	}
	return false
}

func (tc *TokenCache) GetServiceEndpoints(serviceType string, serviceName string) []Endpoint {
	endpoints := []Endpoint{}
	for _, catalog := range tc.token.Catalogs {
		if catalog.Type != serviceType || (serviceName != "" && catalog.Name != serviceName) {
			continue
		}
		return catalog.Endpoints
	}
	return endpoints
}

type Auth struct {
	Identity Identity `json:"identity,omitempty"`
	Scope    Scope    `json:"scope,omitempty"`
}

type AuthBody struct {
	Auth Auth `json:"auth"`
}

type RespToken struct {
	Token         Token `json:"token"`
	XSubjectToken string
}
