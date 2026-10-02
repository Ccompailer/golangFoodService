package interfaces

import (
	"net/http"
)

type IAuthenticator interface {
	Authenticate(next http.Handler) http.Handler
}
