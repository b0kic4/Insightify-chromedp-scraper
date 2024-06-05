package server

import (
	"Insightify-backend/internal/analyze"
	tokenvalidation "Insightify-backend/internal/validateToken"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://insightifyyy.vercel.app/", "https://insightify-backend-3caf92991e4a.herokuapp.com/"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"X-PINGOTHER", "Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}).Handler

	r.Use(corsHandler)

	r.Use(middleware.Logger)
	// NOTE: Protect analysis routes with token validation
	r.Mount("/analysis", analyze.AnalysisRoutes())
	r.Mount("/", s.generalRoutes())

	return r
}

func (s *Server) generalRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", s.HelloWorldHandler)
	r.With(tokenvalidation.TokenAuthMiddleware).Get("/health", s.healthHandler)
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
