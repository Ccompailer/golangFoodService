package identityapi

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
	User         User   `json:"user"`
}

type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"createdAt,omitempty"`
}
