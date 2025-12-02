package ijwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Handler interface {
	SetJWTToken(claims ClaimsParams) (string, error)
	ParseJWTToken(tokenString string) (*Claims, error)
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

func (h *HandlerImpl) ParseJWTToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return h.Secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (h *HandlerImpl) SetJWTToken(claimsParams ClaimsParams) (string, error) {
	claims := Claims{
		ClaimsParams: claimsParams,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.TokenExpire)),
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
