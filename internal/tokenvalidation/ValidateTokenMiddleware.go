package tokenvalidation

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

// TokenClaims extends standard jwt claims
type TokenClaims struct {
	UserID   uint   `json:"userID"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	jwt.StandardClaims
}

// ValidateToken checks if the token is valid and returns the claims
func ValidateToken(tokenString string) (*TokenClaims, error) {
	claims := &TokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Use the secretKey from the environment variable
		secretKey := os.Getenv("JWT_SECRET_KEY")
		if secretKey == "" {
			return nil, fmt.Errorf("JWT_SECRET_KEY is not set")
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid || time.Now().Unix() > claims.ExpiresAt {
		return nil, fmt.Errorf("token is invalid or expired")
	}

	return claims, nil
}
