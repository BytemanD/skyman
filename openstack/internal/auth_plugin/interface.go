package auth_plugin

import (
	"net/http"
	"time"

	"github.com/BytemanD/skyman/openstack/model"
	"github.com/go-resty/resty/v2"
)

type AuthPlugin interface {
	GetToken() (*model.Token, error)
	SetLocalTokenExpire(expire int)
	GetEndpoint(region string, sType string, sName string, sInterface string) (string, error)
	TokenIssue() error
	AuthRequest(req *resty.Request) error
	GetSafeHeader(header http.Header) http.Header
	GetProjectId() (string, error)
	IsAdmin() bool
	// set http client
	SetTimeout(t time.Duration)
	SetRetryCount(c int)
	SetRetryWaitTime(t time.Duration)
	SetRetryMaxWaitTime(t time.Duration)
}
