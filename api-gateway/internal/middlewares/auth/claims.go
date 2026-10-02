package auth

import (
	"slices"

	"github.com/lestrrat-go/jwx/v2/jwt"
)

type Claims struct {
	UserID string
	Roles  []string
	Token  jwt.Token
}

func (c *Claims) HasRole(role string) bool {
	return slices.Contains(c.Roles, role)
}
