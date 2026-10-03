package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/food-service/api-gateway/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Authenticator struct {
	cfg    config.AuthConfig
	keySet jwk.Set
}

type ctxKey struct{}

func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(ctxKey{}).(*Claims)
	return c, ok
}

func ForwardIdentityInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any,
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		if c, ok := ClaimsFromContext(ctx); ok {
			ctx = metadata.AppendToOutgoingContext(ctx,
				"x-user-id", c.UserID,
				"x-user-roles", strings.Join(c.Roles, ","),
			)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func RequireRole(roles ...string) func() gin.HandlerFunc {
	return func() gin.HandlerFunc {
		return func(c *gin.Context) {
			claims, ok := ClaimsFromContext(c.Request.Context())
			if !ok {
				writeAuthError(c, http.StatusUnauthorized, "unauthenticated", "authentication required")
				return
			}
			if slices.ContainsFunc(roles, claims.HasRole) {
				c.Next()
				return
			}
			writeAuthError(c, http.StatusForbidden, "forbidden", "insufficient permissions")
		}
	}
}

func NewAuthenticator(ctx context.Context, cfg config.AuthConfig) (*Authenticator, error) {
	if cfg.Skew == 0 {
		cfg.Skew = 30 * time.Second
	}
	if cfg.RefreshMin == 0 {
		cfg.RefreshMin = 5 * time.Minute
	}

	cache := jwk.NewCache(ctx)
	if err := cache.Register(cfg.JWKSURL, jwk.WithMinRefreshInterval(cfg.RefreshMin)); err != nil {
		return nil, fmt.Errorf("register jwks: %w", err)
	}
	if _, err := cache.Refresh(ctx, cfg.JWKSURL); err != nil {
		return nil, fmt.Errorf("initial jwks fetch: %w", err)
	}

	return &Authenticator{
		cfg:    cfg,
		keySet: jwk.NewCachedSet(cache, cfg.JWKSURL),
	}, nil
}

func (a Authenticator) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, err := bearerToken(c.Request)
		if err != nil {
			writeAuthError(c, http.StatusUnauthorized, "missing_token", err.Error())
			return
		}

		claims, err := a.verify(raw)
		if err != nil {
			writeAuthError(c, http.StatusUnauthorized, "invalid_token", "token is invalid or expired")
			return
		}

		ctx := context.WithValue(c.Request.Context(), ctxKey{}, claims)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func (a Authenticator) verify(raw string) (*Claims, error) {
	opts := []jwt.ParseOption{
		// Ключ ищется по kid в кешированном JWKS. Алгоритм берётся из ключа,
		// а не из заголовка токена, что защищает от подмены alg.
		jwt.WithKeySet(a.keySet, jws.WithInferAlgorithmFromKey(true)),
		jwt.WithValidate(true),
		jwt.WithAcceptableSkew(a.cfg.Skew),
		jwt.WithRequiredClaim(jwt.SubjectKey),
		jwt.WithRequiredClaim(jwt.ExpirationKey),
	}
	if a.cfg.Issuer != "" {
		opts = append(opts, jwt.WithIssuer(a.cfg.Issuer))
	}
	if a.cfg.Audience != "" {
		opts = append(opts, jwt.WithAudience(a.cfg.Audience))
	}

	tok, err := jwt.Parse([]byte(raw), opts...)
	if err != nil {
		return nil, err
	}

	return &Claims{
		UserID: tok.Subject(),
		Roles:  extractRoles(tok),
		Token:  tok,
	}, nil
}

func bearerToken(r *http.Request) (string, error) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", errors.New("authorization header is required")
	}
	const prefix = "Bearer "
	if len(h) <= len(prefix) || !strings.EqualFold(h[:len(prefix)], prefix) {
		return "", errors.New("authorization header must be 'Bearer <token>'")
	}
	return strings.TrimSpace(h[len(prefix):]), nil
}

func extractRoles(tok jwt.Token) []string {
	v, ok := tok.Get("roles")
	if !ok {
		return nil
	}
	switch roles := v.(type) {
	case []any:
		out := make([]string, 0, len(roles))
		for _, r := range roles {
			if s, ok := r.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return roles
	case string:
		return []string{roles}
	}
	return nil
}

func writeAuthError(c *gin.Context, status int, code, msg string) {
	if status == http.StatusUnauthorized {
		c.Header("WWW-Authenticate", `Bearer realm="api"`)
	}

	c.AbortWithStatusJSON(status, gin.H{
		"error": gin.H{"code": code, "message": msg},
	})
}
