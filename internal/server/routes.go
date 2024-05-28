package server

import (
	"Insightify-backend/internal/analyze"
	"Insightify-backend/internal/auth"
	"Insightify-backend/internal/middlewares"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/markbates/goth/gothic"
	"github.com/rs/cors"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // adjust this for production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"X-PINGOTHER", "Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}).Handler

	r.Use(middleware.Logger)
	r.Use(corsHandler)
	r.With(middlewares.VerifyTokenMiddleware).Mount("/analysis", analyze.AnalysisRoutes())
	r.Mount("/", s.generalRoutes())
	r.Mount("/auth", s.authRoutes())
	r.Mount("/authenticated", s.authenticatedRoutes())

	return r
}

func (s *Server) generalRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", s.HelloWorldHandler)
	r.Get("/health", s.healthHandler)
	return r
}

func (s *Server) authRoutes() chi.Router {
	r := chi.NewRouter()
	authHandler := auth.NewAuthHandler(s.dbService.DB())
	r.Post("/signup", authHandler.CreateUserWithCredentials)
	r.Post("/login", authHandler.LoginUserWithCredentials)
	r.Get("/{provider}", auth.GetProviderHandler)
	r.Get("/{provider}/callback", authHandler.SaveUserToDatabaseProvider)
	r.Get("/logout/{provider}", func(w http.ResponseWriter, r *http.Request) {
		gothic.Logout(w, r)
		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusTemporaryRedirect)
	})

	return r
}

func (s *Server) authenticatedRoutes() chi.Router {
	r := chi.NewRouter()
	authHandler := auth.NewAuthHandler(s.dbService.DB())
	r.With(middlewares.VerifyTokenMiddleware).Get("/currentUser", authHandler.GetUserFromToken)
	r.With(middlewares.VerifyTokenMiddleware).Get("/profile", s.HelloWorldHandler)
	r.With(middlewares.VerifyTokenMiddleware).Get("/logout", auth.LogoutHandler)
	return r
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error handling JSON marshal. Err: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(jsonResp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(s.dbService.Health())
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(jsonResp)
}
