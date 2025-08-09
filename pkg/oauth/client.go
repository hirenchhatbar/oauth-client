package oauth

import (
	"github.com/hirenchhatbar/oauth-client/internal/oauth"
)

func Init() error {
	return oauth.Init()
}

func Listen() error {
	return oauth.Listen()
}

func GetToken() (string, error) {
	return oauth.GetToken()
}
