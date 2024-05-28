package auth

import (
	"Insightify-backend/internal/database/models"
	"Insightify-backend/internal/middlewares"
	"Insightify-backend/internal/services"
	"Insightify-backend/internal/tokenvalidation"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/markbates/goth/gothic"
	"golang.org/x/crypto/bcrypt"
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

func (h *AuthHandler) GetUserFromToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
		return
	}

	claims, err := tokenvalidation.ValidateToken(cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	identifier := ""
	if claims.Email != "" {
		identifier = claims.Email
	} else if claims.Username != "" {
		identifier = claims.Username
	}

	user, err := h.GetUserByIdentifier(identifier)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "Unauthorized: User not found", http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"ID":            user.ID,
		"Username":      user.Username,
		"FullName":      user.FullName,
		"Email":         user.Email,
		"Provider":      user.Provider,
		"ProviderID":    user.ProviderID,
		"AvatarURL":     user.AvatarURL,
		"VerifiedEmail": user.VerifiedEmail,
	}
	json.NewEncoder(w).Encode(response)
}

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

func (h *AuthHandler) CreateUserWithCredentials(w http.ResponseWriter, r *http.Request) {
	var creds SignUpUserCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Use the passed-in *gorm.DB instance
	user, err := h.GetUserByIdentifier(creds.Email)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if user != nil {
		http.Error(w, "Account with that email already exists. Please login.", http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash the password", http.StatusInternalServerError)
		return
	}

	newUser := models.User{
		Username:      creds.Username,
		PasswordHash:  string(hashedPassword),
		Email:         creds.Email,
		FullName:      creds.FullName,
		VerifiedEmail: false,
	}

	// Correctly use the db instance to call Create
	if err := h.DB.Create(&newUser).Error; err != nil {
		log.Printf("Failed to create user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	tokenString, err := middlewares.CreateToken(newUser)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    tokenString,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token":       tokenString,
		"redirectURL": "http://localhost:3000",
	})
}

func (h *AuthHandler) LoginUserWithCredentials(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Retrieve user by email
	user, err := h.GetUserByIdentifier(creds.Email)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Email does not exist."})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid password."})
		return
	}

	tokenString, err := middlewares.CreateToken(*user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    tokenString,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
		Secure:   true, // Remember to set to true in production for HTTPS
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":       tokenString,
		"redirectURL": "http://localhost:3000",
	})
}

func (h *AuthHandler) SaveUserToDatabaseProvider(w http.ResponseWriter, r *http.Request) {
	ctxProvider := r.Context().Value("provider")
	provider, ok := ctxProvider.(string)
	if !ok || provider == "" {
		provider := chi.URLParam(r, "provider")
		if provider == "" {
			http.Error(w, "Provider not specified", http.StatusBadRequest)
			return
		}
	}

	oauthUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		log.Fatal("Error has ocurred")
	}

	newUser := models.User{
		Email:         oauthUser.Email,
		Username:      oauthUser.NickName,
		FullName:      oauthUser.Name,
		Provider:      oauthUser.Provider,
		ProviderID:    oauthUser.UserID,
		AvatarURL:     oauthUser.AvatarURL,
		VerifiedEmail: true,
	}

	ctx := r.Context()

	userService := services.NewUserService(h.DB)

	user, err := userService.CreateUserOrUpdate(ctx, newUser)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	tokenString, err := middlewares.CreateToken(*user)
	if err != nil {
		log.Printf("Failed to generate token: %v\n", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    tokenString,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "http://localhost:3000", http.StatusFound)
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
