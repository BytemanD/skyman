package auth_plugin

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/session"
	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
)

const (
	DEFAUL_REGION  = "RegionOne"
	TYPE_COMPUTE   = "compute"
	TYPE_VOLUME    = "volume"
	TYPE_VOLUME_V2 = "volumev2"
	TYPE_VOLUME_V3 = "volumev3"
	TYPE_IDENTITY  = "identity"
	TYPE_IMAGE     = "image"
	TYPE_NETWORK   = "network"

	INTERFACE_PUBLIC   = "public"
	INTERFACE_ADMIN    = "admin"
	INTERFACE_INTERVAL = "internal"

	URL_AUTH_TOKEN = "/auth/tokens"

	X_SUBNECT_TOKEN = "X-Subject-Token"
	X_AUTH_TOKEN    = "X-Auth-Token"
)

type PasswordAuthPlugin struct {
	AuthUrl           string
	Username          string
	Password          string
	ProjectName       string
	UserDomainName    string
	ProjectDomainName string

	LocalTokenExpireSecond int
	token                  *model.Token
	expiredAt              time.Time

	tokenLock *sync.Mutex

	session *resty.Client
}

func (plugin *PasswordAuthPlugin) SetLocalTokenExpire(expireSeconds int) {
	plugin.LocalTokenExpireSecond = expireSeconds
}

func (plugin *PasswordAuthPlugin) SetTimeout(timeout time.Duration) {
	plugin.session.SetTimeout(timeout)
}
func (plugin *PasswordAuthPlugin) SetRetryCount(c int) {
	plugin.session.SetRetryCount(c)
}
func (plugin *PasswordAuthPlugin) SetRetryWaitTime(t time.Duration) {
	plugin.session.SetRetryWaitTime(t)
}
func (plugin *PasswordAuthPlugin) SetRetryMaxWaitTime(t time.Duration) {
	plugin.session.SetRetryMaxWaitTime(t)
}

func (plugin *PasswordAuthPlugin) IsTokenExpired() bool {
	if plugin.token == nil {
		return true
	}
	if plugin.expiredAt.Before(time.Now()) {
		console.Warn("token exipred, expired at: %s , now: %s", plugin.expiredAt, time.Now())
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

func (plugin *PasswordAuthPlugin) GetToken() (*model.Token, error) {
	if err := plugin.makesureTokenValid(); err != nil {
		return nil, err
	}
	return plugin.token, nil
}

type AuthBody struct {
	Auth model.Auth `json:"auth"`
}

func (client PasswordAuthPlugin) newAuthReqBody() AuthBody {
	authData := model.Auth{
		Identity: model.Identity{
			Methods: []string{"password"},
			Password: model.Password{
				User: model.User{
					Name: client.Username, Password: client.Password,
					Domain: model.Domain{Name: client.UserDomainName}}},
		},
		Scope: model.Scope{Project: model.Project{
			Name:   client.ProjectName,
			Domain: model.Domain{Name: client.ProjectDomainName}},
		},
	}
	return AuthBody{Auth: authData}
}

func (plugin *PasswordAuthPlugin) TokenIssue() error {
	respBody := struct {
		Token model.Token `json:"token"`
	}{}
	resp, err := plugin.session.R().SetBody(plugin.newAuthReqBody()).
		SetResult(&respBody).Post(URL_AUTH_TOKEN)

	if err != nil {
		return fmt.Errorf("token issue failed, %s", err)
	}
	plugin.token = &respBody.Token
	plugin.token.TokenId = resp.Header().Get(X_SUBNECT_TOKEN)
	plugin.expiredAt = time.Now().Add(time.Second * time.Duration(plugin.LocalTokenExpireSecond))
	return nil
}

func (plugin *PasswordAuthPlugin) GetEndpoint(region string, sType string, sName string, sInterface string) (string, error) {
	if err := plugin.makesureTokenValid(); err != nil {
		return "", fmt.Errorf("get token failed: %w", err)
	}
	if region == "" {
		return "", fmt.Errorf("region is required")
	}
	for _, catalog := range plugin.token.Catalogs {
		if catalog.Type != sType || (sName != "" && catalog.Name != sName) {
			continue
		}
		for _, endpoint := range catalog.Endpoints {
			if endpoint.Interface == sInterface && endpoint.Region == region {
				return endpoint.Url, nil
			}
		}
	}
	return "", fmt.Errorf("endpoint %s:%s:%s for region '%s' not found",
		sType, sName, sInterface, region)
}

func (plugin PasswordAuthPlugin) AuthRequest(req *resty.Request) error {
	token, err := plugin.GetToken()
	if err != nil {
		return err
	}
	if req.Header.Get(X_AUTH_TOKEN) == token.TokenId {
		return nil
	}
	console.Debug("set header %s: %s", X_AUTH_TOKEN, token.TokenId)
	req.Header.Set(X_AUTH_TOKEN, token.TokenId)
	return nil
}
func (plugin PasswordAuthPlugin) GetSafeHeader(header http.Header) http.Header {
	safeHeaders := http.Header{}
	for k, v := range header {
		if k == X_AUTH_TOKEN {
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
func (plugin *PasswordAuthPlugin) Roles() []string {
	plugin.makesureTokenValid()
	if plugin.token == nil {
		return []string{}
	}
	return lo.Map(plugin.token.Roles, func(item model.Role, _ int) string {
		return item.Name
	})
}
func (plugin *PasswordAuthPlugin) IsAdmin() bool {
	return lo.Contains(plugin.Roles(), "admin")
}

func NewPasswordAuthPlugin(authUrl string, user model.User, project model.Project) *PasswordAuthPlugin {
	plugin := &PasswordAuthPlugin{
		session: session.DefaultRestyClient(strings.TrimSuffix(authUrl, "/")).
			OnBeforeRequest(session.LogBeforeRequest),
		AuthUrl:           strings.TrimSuffix(authUrl, "/"),
		Username:          user.Name,
		Password:          user.Password,
		UserDomainName:    user.Domain.Name,
		ProjectName:       project.Name,
		ProjectDomainName: project.Domain.Name,
		tokenLock:         &sync.Mutex{},
	}
	return plugin
}
