package oauth

import (
	"github.com/hirenchhatbar/oauth-client/internal/oauth"
)

func Listen(port int) error {
	return oauth.Listen(port)
}
