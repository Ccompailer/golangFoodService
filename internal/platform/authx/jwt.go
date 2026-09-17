package authx

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"golangFoodService/internal/platform/apperr"
	"golangFoodService/internal/platform/id"
)

type Role string

const (
	RoleCustomer   Role = "customer"
	RoleRestaurant Role = "restaurant"
	RoleCourier    Role = "courier"
)

func ParseRole(v string) (Role, error) {
	switch Role(v) {
	case RoleCustomer, RoleRestaurant, RoleCourier:
		return Role(v), nil
	default:
		return "", apperr.Invalid("unknown role")
	}
}

type Claims struct {
	Email     string `json:"email"`
	Role      Role   `json:"role"`
	TokenType string `json:"typ"`
	jwt.RegisteredClaims
}

type Tokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

type JWT struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewJWT(secret string, accessTTL, refreshTTL time.Duration) *JWT {
	return &JWT{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (j *JWT) Issue(userID, email string, role Role) (Tokens, error) {
	now := time.Now().UTC()
	access, err := j.sign(userID, email, role, "access", now.Add(j.accessTTL))
	if err != nil {
		return Tokens{}, err
	}
	refresh, err := j.sign(userID, email, role, "refresh", now.Add(j.refreshTTL))
	if err != nil {
		return Tokens{}, err
	}
	return Tokens{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(j.accessTTL.Seconds()),
	}, nil
}

func (j *JWT) ParseAccess(token string) (*Claims, error) {
	return j.parse(token, "access")
}

func (j *JWT) ParseRefresh(token string) (*Claims, error) {
	return j.parse(token, "refresh")
}

func (j *JWT) sign(userID, email string, role Role, tokenType string, exp time.Time) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		Email:     email,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "golangFoodService",
			ID:        id.New(),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(j.secret)
}

func (j *JWT) parse(token, wantType string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, apperr.Unauthorized("invalid token")
		}
		return j.secret, nil
	})
	if err != nil || parsed == nil || !parsed.Valid {
		return nil, apperr.Unauthorized("invalid token")
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || claims.TokenType != wantType || claims.Subject == "" {
		return nil, apperr.Unauthorized("invalid token")
	}
	return claims, nil
}
