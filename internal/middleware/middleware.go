package middleware

import (
	"fmt"
	"net/http"
	"os"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

func TokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := authHeader[len("Bearer "):]

		tokenIssuer := fmt.Sprintf("https://%v", os.Getenv("KINDE_ENVIRONMENT_DOMAIN"))
		apiAudience := os.Getenv("MY_API_AUDIENCE")

		jwksURL := fmt.Sprintf("%v/.well-known/jwks", tokenIssuer)
		jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{})
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		parsedToken, err := jwt.Parse(tokenString, jwks.Keyfunc,
			jwt.WithValidMethods([]string{"RS256"}), // verifying the signing algorithm
			jwt.WithIssuer(tokenIssuer),             // verifying the token issuer
			jwt.WithAudience(apiAudience))           // verifying that the token is for correct audience
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if !parsedToken.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
