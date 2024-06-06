package tokenvalidation

import (
	"fmt"
	"net/http"
	"os"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

// TokenAuthMiddleware validates JWT tokens passed via the Authorization header or query parameter
func TokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string

		// Check for token in Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			// Check for token in query parameter
			tokenString = r.URL.Query().Get("token")
		}

		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		err := validateToken(tokenString)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unauthorized: %v", err), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// validateToken verifies the provided JWT token without checking the audience
func validateToken(tokenString string) error {
	tokenIssuer := os.Getenv("KINDE_ENVIRONMENT_DOMAIN")

	if tokenIssuer == "" {
		return fmt.Errorf("missing environment variable KINDE_ENVIRONMENT_DOMAIN")
	}

	jwksURL := fmt.Sprintf("%v/.well-known/jwks", tokenIssuer)
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{})
	if err != nil {
		fmt.Println("unauthorized1: ", err)
		return fmt.Errorf("unauthorized: %v", err)
	}

	parsedToken, err := jwt.Parse(tokenString, jwks.Keyfunc,
		jwt.WithValidMethods([]string{"RS256"}), // verifying the signing algorithm
		jwt.WithIssuer(tokenIssuer))             // verifying the token issuer
	if err != nil {
		fmt.Println("unauthorized2: ", err)
		return fmt.Errorf("unauthorized: %v", err)
	}

	// Check if token is valid
	if !parsedToken.Valid {
		fmt.Println("unauthorized3: ", err)
		return fmt.Errorf("unauthorized: invalid token")
	}

	return nil
}
