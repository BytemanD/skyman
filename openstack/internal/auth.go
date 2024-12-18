package internal

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack/auth"
	"github.com/go-resty/resty/v2"
)

const (
	TYPE_COMPUTE   string = "compute"
	TYPE_VOLUME    string = "volume"
	TYPE_VOLUME_V2 string = "volumev2"
	TYPE_VOLUME_V3 string = "volumev3"
	TYPE_IDENTITY  string = "identity"
	TYPE_IMAGE     string = "image"
	TYPE_NETWORK   string = "network"

	INTERFACE_PUBLIC   string = "public"
	INTERFACE_ADMIN    string = "admin"
	INTERFACE_INTERVAL string = "internal"

	URL_AUTH_TOKEN string = "/auth/tokens"
)

type PasswordAuthPlugin struct {
	AuthUrl           string
	Username          string
	Password          string
	ProjectName       string
	UserDomainName    string
	ProjectDomainName string
	RegionName        string

	LocalTokenExpireSecond int
	token                  *auth.Token
	tokenId                string
	expiredAt              time.Time

	tokenLock *sync.Mutex

	session *resty.Client
}

func (plugin PasswordAuthPlugin) Region() string {
	return plugin.RegionName
}

func (plugin *PasswordAuthPlugin) SetRegion(region string) {
	plugin.RegionName = region
}

func (plugin *PasswordAuthPlugin) SetLocalTokenExpire(expireSeconds int) {
	plugin.LocalTokenExpireSecond = expireSeconds
}

func (plugin *PasswordAuthPlugin) IsTokenExpired() bool {
	if plugin.tokenId == "" {
		return true
	}
	if plugin.expiredAt.Before(time.Now()) {
		logging.Warning("token exipred, expired at: %s , now: %s", plugin.expiredAt, time.Now())
		return true
	}
	return false
}

func (plugin *PasswordAuthPlugin) makesureTokenValid() error {
	plugin.tokenLock.Lock()
	defer plugin.tokenLock.Unlock()

	if plugin.IsTokenExpired() {
		return plugin.TokenIssue()
	}
	return nil
}

func (plugin *PasswordAuthPlugin) GetToken() (*auth.Token, error) {
	plugin.makesureTokenValid()
	return plugin.token, nil
}
func (plugin *PasswordAuthPlugin) GetTokenId() (string, error) {
	if err := plugin.makesureTokenValid(); err != nil {
		return "", err
	}
	return plugin.tokenId, nil
}

type AuthBody struct {
	Auth auth.Auth `json:"auth"`
}

func (client PasswordAuthPlugin) newAuthReqBody() AuthBody {
	authData := auth.Auth{
		Identity: auth.Identity{
			Methods: []string{"password"},
			Password: auth.Password{
				User: auth.User{
					Name: client.Username, Password: client.Password,
					Domain: auth.Domain{Name: client.UserDomainName}}},
		},
		Scope: auth.Scope{Project: auth.Project{
			Name:   client.ProjectName,
			Domain: auth.Domain{Name: client.ProjectDomainName}},
		},
	}
	return AuthBody{Auth: authData}
}

func (plugin *PasswordAuthPlugin) TokenIssue() error {
	respBody := struct {
		Token auth.Token `json:"token"`
	}{}
	resp, err := plugin.session.R().SetBody(plugin.newAuthReqBody()).
		SetResult(&respBody).
		Post(fmt.Sprintf("%s%s", plugin.AuthUrl, URL_AUTH_TOKEN))
	if err != nil || resp.Error() != nil {
		return fmt.Errorf("token issue failed, %s %s", err, resp.Error())
	}
	plugin.tokenId = resp.Header().Get("X-Subject-Token")
	plugin.token = &respBody.Token
	plugin.expiredAt = time.Now().Add(time.Second * time.Duration(plugin.LocalTokenExpireSecond))
	return nil
}
func (plugin *PasswordAuthPlugin) GetServiceEndpoints(sType string, sName string) ([]auth.Endpoint, error) {
	token, err := plugin.GetToken()
	if err != nil {
		return nil, err
	}

	for _, catalog := range token.Catalogs {
		if catalog.Type != sType || (sName != "" && catalog.Name != sName) {
			continue
		}
		return catalog.Endpoints, nil
	}
	return []auth.Endpoint{}, nil
}
func (plugin *PasswordAuthPlugin) GetServiceEndpoint(sType string, sName string, sInterface string) (string, error) {
	if err := plugin.makesureTokenValid(); err != nil {
		return "", fmt.Errorf("get catalogs failed: %s", err)
	}

	for _, catalog := range plugin.token.Catalogs {
		if catalog.Type != sType || (sName != "" && catalog.Name != sName) {
			continue
		}
		for _, endpoint := range catalog.Endpoints {
			if endpoint.Interface == sInterface && endpoint.Region == plugin.RegionName {
				return endpoint.Url, nil
			}
		}
	}
	return "", fmt.Errorf("endpoint %s:%s:%s for region '%s' not found",
		sType, sName, sInterface, plugin.RegionName)
}
func (plugin *PasswordAuthPlugin) SetHttpTimeout(timeout int) *PasswordAuthPlugin {
	plugin.session.SetTimeout(time.Second * time.Duration(timeout))
	return plugin
}
func (plugin *PasswordAuthPlugin) SetRetryWaitTime(waitTime int) *PasswordAuthPlugin {
	plugin.session.SetRetryWaitTime(time.Second * time.Duration(waitTime))
	return plugin
}
func (plugin *PasswordAuthPlugin) SetRetryCount(count int) *PasswordAuthPlugin {
	plugin.session.SetRetryCount(count)
	return plugin
}

func (plugin PasswordAuthPlugin) AuthRequest(req *resty.Request) error {
	tokenId, err := plugin.GetTokenId()
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", tokenId)
	return nil
}
func (plugin PasswordAuthPlugin) GetSafeHeader(header http.Header) http.Header {
	safeHeaders := http.Header{}
	for k, v := range header {
		if k == "X-Auth-Token" {
			safeHeaders[k] = []string{"<TOKEN>"}
		} else {
			safeHeaders[k] = v
		}
	}
	return safeHeaders
}
func (plugin PasswordAuthPlugin) GetProjectId() (string, error) {
	if err := plugin.makesureTokenValid(); err != nil {
		return "", err
	}
	return plugin.token.Project.Id, nil
}
func (plugin PasswordAuthPlugin) IsAdmin() bool {
	for _, role := range plugin.token.Roles {
		if role.Name == "admin" {
			return true
		}
	}
	return false
}
func NewPasswordAuth(authUrl string, user auth.User, project auth.Project, regionName string) *PasswordAuthPlugin {
	return &PasswordAuthPlugin{
		session: resty.New().OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
			logging.Debug("REQ: %s %s, Query: %v", r.Method, r.URL, r.QueryParam.Encode())
			return nil
		}),
		AuthUrl:           authUrl,
		Username:          user.Name,
		Password:          user.Password,
		UserDomainName:    user.Domain.Name,
		ProjectName:       project.Name,
		ProjectDomainName: project.Domain.Name,
		RegionName:        regionName,
		tokenLock:         &sync.Mutex{},
	}
}
