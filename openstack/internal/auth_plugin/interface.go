package auth_plugin

import (
	"net/http"

	"github.com/BytemanD/skyman/openstack/model"
	"github.com/go-resty/resty/v2"
)

type AuthPlugin interface {
	GetToken() (*model.Token, error)
	GetTokenId() (string, error)
	SetLocalTokenExpire(expire int)
	GetServiceEndpoint(sType string, sName string, sInterface string) (string, error)
	TokenIssue() error
	Region() string
	SetRegion(region string)
	AuthRequest(req *resty.Request) error
	GetSafeHeader(header http.Header) http.Header
	GetProjectId() (string, error)
	IsAdmin() bool
}
