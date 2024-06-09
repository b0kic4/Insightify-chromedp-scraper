package scraper

import (
	"Insightify-backend/internal/utils"
	"context"
	"fmt"

	"firebase.google.com/go/storage"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type Scraper struct {
	FirebaseStorage *storage.Client
	RedisClient     *redis.Client
}

type WebSocketMessage struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

func NewScraper(ctx context.Context) *Scraper {
	storage := utils.NewFirebaseClient(ctx)
	return &Scraper{
		FirebaseStorage: storage,
	}
}

func (s *Scraper) CaptureAndUpload(url string, conn *websocket.Conn) []string {
	ctx, cancel, err := s.navigateAndSetup(url, conn)
	if err != nil {
		s.sendWebSocketMessage(conn, WebSocketMessage{Type: "error", Content: "Failed to setup navigation"})
		return nil
	}
	defer cancel()

	s.sendWebSocketMessage(conn, WebSocketMessage{Type: "status", Content: "Navigation to the provided url completed"})

	lastScrollY, err := s.determineHeight(ctx)
	if err != nil {
		s.sendWebSocketMessage(conn, WebSocketMessage{Type: "error", Content: "Failed to determine page height"})
		return nil
	}
	fmt.Println("lastScrollY: ", lastScrollY)

	screenshots := s.captureScreenshots(conn, ctx, lastScrollY)
	if len(screenshots) > 0 {
		s.sendWebSocketMessage(conn, WebSocketMessage{Type: "images", Content: screenshots})
	} else {
		s.sendWebSocketMessage(conn, WebSocketMessage{Type: "error", Content: "No screenshots were captured"})
	}
	return screenshots
}
