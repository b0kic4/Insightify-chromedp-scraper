package server

import (
	"Insightify-backend/internal/database"
	"Insightify-backend/internal/services"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port        int
	dbService   database.Service
	userService *services.UserService
}

func NewServer() *http.Server {
	// Convert the PORT environment variable from string to int
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		fmt.Printf("Error converting PORT: %s\n", err)
		return nil
	}

	// Initialize the database service
	dbService := database.New()

	// Pass the GORM DB from the database service to the UserService
	userService := services.NewUserService(dbService.DB())

	// Create the server struct
	server := &Server{
		port:        port,
		dbService:   dbService,
		userService: userService,
	}

	// Configure the HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", server.port),
		Handler:      server.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return httpServer
}
