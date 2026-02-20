package jwt

import (
	"errors"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
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

func (h *HandlerImpl) SetJWTToken(claimsParams ClaimsParams) (string, error) {
	claims := Claims{
		ClaimsParams: claimsParams,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(h.TokenExpire * time.Second)),
		},
	}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, &claims)
	return token.SignedString(h.Secret)
}

func (h *HandlerImpl) ParseJWTToken(tokenString string) (*Claims, error) {
	token, err := jwtv5.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwtv5.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwtv5.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return h.Secret, nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

type Claims struct {
	ClaimsParams
	jwtv5.RegisteredClaims
}

type ClaimsParams struct {
	UserID     uint64 `json:"user_id"`
	UserRole   string `json:"user_role"`
	UserStatus int64  `json:"user_status"`
}
