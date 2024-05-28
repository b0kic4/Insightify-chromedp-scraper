package main

import (
	"Insightify-backend/internal/server"
	"fmt"
)

func main() {
	server := server.NewServer()
	// auth.NewAuth()
	err := server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
