package auth

import (
	"Insightify-backend/internal/database/models"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/markbates/goth/gothic"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{DB: db}
}

// when user is singin up with credentials
type SignUpUserCredentials struct {
	Username        string `json:"username"`
	FullName        string `json:"fullName"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

func (h *AuthHandler) GetUserByIdentifier(identifier string) (*models.User, error) {
	var user models.User
	err := h.DB.Where("email = ? OR username = ?", identifier, identifier).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No user found is not an error in this context; return nil user and no error
			return nil, nil
		}
		// Actual error accessing database
		return nil, err
	}
	return &user, nil
}

// func (h *AuthHandler) GetUserFromToken(w http.ResponseWriter, r *http.Request) {
// 	cookie, err := r.Cookie("jwt_token")
// 	if err != nil {
// 		http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
// 		return
// 	}
//
// 	claims, err := tokenvalidation.ValidateToken(cookie.Value)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusUnauthorized)
// 		return
// 	}
//
// 	identifier := ""
// 	if claims.Email != "" {
// 		identifier = claims.Email
// 	} else if claims.Username != "" {
// 		identifier = claims.Username
// 	}
//
// 	user, err := h.GetUserByIdentifier(identifier)
// 	if err != nil {
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 		return
// 	}
// 	if user == nil {
// 		http.Error(w, "Unauthorized: User not found", http.StatusUnauthorized)
// 		return
// 	}
//
// 	response := map[string]interface{}{
// 		"ID":            user.ID,
// 		"Username":      user.Username,
// 		"FullName":      user.FullName,
// 		"Email":         user.Email,
// 		"Provider":      user.Provider,
// 		"ProviderID":    user.ProviderID,
// 		"AvatarURL":     user.AvatarURL,
// 		"VerifiedEmail": user.VerifiedEmail,
// 	}
// 	json.NewEncoder(w).Encode(response)
// }

func GetProviderHandler(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")

	session, err := gothic.Store.Get(r, "gothic-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["goth_provider"] = provider
	if err := session.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gothic.BeginAuthHandler(w, r)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}
