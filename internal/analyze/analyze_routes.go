package analyze

import (
	"github.com/go-chi/chi/v5"
)

func AnalysisRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/ws", WebSocketHandler)
	return r
}
