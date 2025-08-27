package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
	expectedAlg   string
}

func NewJWTManager(secretKey string, duration time.Duration, alg string,
) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		tokenDuration: duration,
		expectedAlg:   alg,
	}
}

type UserClaims struct {
	jwt.RegisteredClaims
	UserID int64 `json:"user_id"`
}

func (j *JWTManager) Generate(userID int64) (string, error) {
	claims := &UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

func (j *JWTManager) Verify(tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&UserClaims{},
		func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != j.expectedAlg {
				return nil, fmt.Errorf("unexpected signing method: %v",
					token.Header["alg"])
			}
			return []byte(j.secretKey), nil
		})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}
