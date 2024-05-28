package scraper

import (
	"Insightify-backend/internal/utils"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

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

type CachedData struct {
	Screenshots []string `json:"screenshots"`
	AIResponse  string   `json:"ai_response"`
}

func NewScraper(ctx context.Context) *Scraper {
	storage := utils.NewFirebaseClient(ctx)
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADD"),
		Password: os.Getenv("REDIS_PW"),
		DB:       0,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	fmt.Println("Connected to Redis successfully")

	return &Scraper{
		FirebaseStorage: storage,
		RedisClient:     rdb,
	}
}

func (s *Scraper) CaptureAndUpload(url string, market string, audience string, insight string, bucketName string, conn *websocket.Conn) []string {
	fmt.Println("Market:", market)
	fmt.Println("Audience:", audience)
	fmt.Println("Insight:", insight)

	var cachedData CachedData
	cachedResult, err := s.RedisClient.Get(context.Background(), url).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(cachedResult), &cachedData); err == nil {
			s.sendWebSocketMessage(conn, WebSocketMessage{Type: "images", Content: cachedData.Screenshots})
			s.sendWebSocketMessage(conn, WebSocketMessage{Type: "ai_response", Content: cachedData.AIResponse})
			return cachedData.Screenshots
		}
	}

	// No valid cache found, proceed to capture and analyze
	ctx, cancel, err := s.navigateAndSetup(url, conn)
	if err != nil {
		s.sendWebSocketMessage(conn, WebSocketMessage{Type: "error", Content: "Failed to setup navigation"})
		return nil
	}
	defer cancel()

	lastScrollY, err := s.determineHeight(ctx)
	if err != nil {
		s.sendWebSocketMessage(conn, WebSocketMessage{Type: "error", Content: "Failed to determine page height"})
		return nil
	}
	fmt.Println("lastScrollY: ", lastScrollY)
	// NOTE: html, screenshots = s.=||=
	_, screenshots := s.captureScreenshotsAndExtractCode(conn, ctx, bucketName, lastScrollY)
	if len(screenshots) > 0 {
		s.sendWebSocketMessage(conn, WebSocketMessage{Type: "images", Content: screenshots})

		// Generate AI response
		aiResponse, err := s.generateAIResponse(screenshots, market, audience, insight)
		if err != nil {
			s.sendWebSocketMessage(conn, WebSocketMessage{Type: "error", Content: "Error processing the AI request"})
			return screenshots
		}

		s.sendWebSocketMessage(conn, WebSocketMessage{Type: "ai_response", Content: aiResponse})
		fmt.Println("ai response: ", aiResponse)

		s.cacheDataInRedis(url, screenshots, aiResponse)
	} else {
		s.sendWebSocketMessage(conn, WebSocketMessage{Type: "error", Content: "No screenshots were captured"})
	}
	return screenshots
}
