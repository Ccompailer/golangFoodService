package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/food-service/api-gateway/internal/config"
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

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, ok := ClaimsFromContext(r.Context())
			if !ok {
				writeAuthError(w, http.StatusUnauthorized, "unauthenticated", "authentication required")
				return
			}
			if slices.ContainsFunc(roles, c.HasRole) {
				next.ServeHTTP(w, r)
				return
			}
			writeAuthError(w, http.StatusForbidden, "forbidden", "insufficient permissions")
		})
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

func (a Authenticator) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := bearerToken(r)
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, "missing_token", err.Error())
			return
		}

		claims, err := a.verify(raw)
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, "invalid_token", "token is invalid or expired")
			return
		}

		ctx := context.WithValue(r.Context(), ctxKey{}, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
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

func writeAuthError(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	if status == http.StatusUnauthorized {
		w.Header().Set("WWW-Authenticate", `Bearer realm="api"`)
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": code, "message": msg},
	})
}
