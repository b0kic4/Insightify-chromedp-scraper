package auth

import (
	"log"
	"os"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
)

func NewAuth() {
	err := godotenv.Load()
	const (
		IsProd = false
		MaxAge = 86400 * 30
	)
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		log.Fatal("SESSION_SECRET environment variable is not set")
	}

	store := sessions.NewCookieStore([]byte(sessionSecret))
	store.MaxAge(MaxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = IsProd

	gothic.Store = store

	goth.UseProviders(google.New(
		os.Getenv("GOOGLE_KEY"),
		os.Getenv("GOOGLE_SECRET"),
		"http://localhost:4000/auth/google/callback"),
		github.New(os.Getenv("GITHUB_KEY"), os.Getenv("GITHUB_SECRET"), "http://localhost:4000/auth/github/callback"))
}
