package auth

import (
	"net/http"

	"github.com/go-resty/resty/v2"
)

type AuthPlugin interface {
	GetToken() (*Token, error)
	GetTokenId() (string, error)
	SetLocalTokenExpire(expire int)
	GetServiceEndpoint(sType string, sName string, sInterface string) (string, error)
	TokenIssue() error
	Region() string
	AuthRequest(req *resty.Request) error
	GetSafeHeader(header http.Header) http.Header
	GetProjectId() (string, error)
}
