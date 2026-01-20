package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Handler interface {
	SetJWTToken(claims ClaimsParams) (string, error)
}

type HandlerImpl struct {
	Secret      []byte
	TokenExpire time.Duration
}

func NewHandler(secret string, expire time.Duration) Handler {
	return &HandlerImpl{
		Secret:      []byte(secret),
		TokenExpire: expire,
	}
}

func (h *HandlerImpl) SetJWTToken(claimsParams ClaimsParams) (string, error) {
	claims := Claims{
		ClaimsParams: claimsParams,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.TokenExpire * time.Second)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	return token.SignedString(h.Secret)
}

type Claims struct {
	ClaimsParams
	jwt.RegisteredClaims
}

type ClaimsParams struct {
	UserId     uint64 `json:"user_id"`
	UserRole   string `json:"user_role"`
	UserStatus int64  `json:"user_status"`
}
