package middlewares

import (
	"Insightify-backend/internal/database/models"
	"Insightify-backend/internal/tokenvalidation"
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = os.Getenv("JWT_SECRET_KEY")

type contextKey string

func CreateToken(user models.User) (string, error) {
	// Define the expiration time for the token
	expirationTime := time.Now().Add(7 * 24 * time.Hour) // 7 days for token to be valid

	// Initialize the claims with fields common to all providers
	claims := jwt.MapClaims{
		"exp":    expirationTime.Unix(),
		"userid": user.ID,
	}

	switch user.Provider {
	case "github":
		claims["username"] = user.Username
	case "google", "":
		claims["email"] = user.Email
	}

	// Create the JWT token with the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		fmt.Printf("Failed to sign token: %v\n", err)
		return "", err
	}

	return tokenString, nil
}

func VerifyTokenMiddleware(next http.Handler) http.Handler {
	userClaimsKey := contextKey("userClaims")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("jwt_token")
		if err != nil {
			if err == http.ErrNoCookie {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tokenString := cookie.Value

		// Validate the token and obtain claims
		claims, err := tokenvalidation.ValidateToken(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userClaimsKey, claims)

		//
		// claims, ok := r.Context().Value(userClaimsKey).(*tokenvalidation.TokenClaims)
		// if !ok {
		//     // Handle the error, e.g., by logging or returning an error response
		//     return
		// }

		// Token is valid, call the next handler with the new context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
