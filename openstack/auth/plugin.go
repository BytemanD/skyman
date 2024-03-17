package auth

import "net/http"

type AuthPlugin interface {
	GetToken() (*Token, error)
	SetTokenExpireSecond(expire int)
	GetAuthTokenId() string
	GetTokenId() (string, error)
	GetServiceEndpoint(sType string, sName string, sInterface string) (string, error)
	TokenIssue() error
	AuthRequest(req *http.Request) error
	Region() string
}
