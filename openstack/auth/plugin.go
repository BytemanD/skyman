package auth

import "net/http"

type AuthPlugin interface {
	GetToken() (*Token, error)
	GetTokenId() (string, error)
	SetLocalTokenExpire(expire int)
	GetServiceEndpoint(sType string, sName string, sInterface string) (string, error)
	TokenIssue() error
	AuthRequest(req *http.Request) error
	Region() string
	GetSafeHeader(header http.Header) http.Header
}
