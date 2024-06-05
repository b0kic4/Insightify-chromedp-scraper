package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/gorilla/websocket"
)

func (s *Scraper) scrollToTop(ctx context.Context) error {
	// Use chromedp to scroll to the top of the page
	err := chromedp.Run(ctx,
		chromedp.KeyEvent(kb.Home),
		chromedp.Sleep(900*time.Millisecond),
		chromedp.KeyEvent(kb.PageUp),
	)
	if err != nil {
		fmt.Println("Failed to scroll back to the top:", err)
		return err
	}
	return nil
}

func (s *Scraper) determineHeight(ctx context.Context) (int, error) {
	var lastScrollY float64 // Use float64 to initially receive the JavaScript floating point value
	// Scroll to the bottom of the page to capture the scroll height
	err := chromedp.Run(ctx,
		chromedp.KeyEvent(kb.End),
		chromedp.Sleep(900*time.Millisecond),
		chromedp.KeyEvent(kb.PageDown),
		chromedp.Evaluate(`Math.round(window.scrollY)`, &lastScrollY), // Use Math.round to round the scroll position
	)
	if err != nil {
		fmt.Println("Failed to scroll to the bottom:", err)
		return 0, err
	}

	// Convert float64 to int after rounding in the JavaScript executed above
	scrollYInt := int(lastScrollY)
	fmt.Println("last scroll: ", scrollYInt)

	// Call scrollToTop to move the page to the top
	err = s.scrollToTop(ctx)
	if err != nil {
		return 0, err // Error from scrollToTop is handled here
	}
	// Optionally, check the scroll position after scrolling up
	var currentScrollY float64
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`Math.round(window.scrollY)`, &currentScrollY),
	)
	if err != nil {
		fmt.Println("Failed to retrieve current scroll position after scrolling to the top:", err)
		return 0, err
	}
	currentScrollYInt := int(currentScrollY)
	fmt.Println("current scroll y: ", currentScrollYInt)

	if currentScrollYInt > 0 {
		fmt.Println("Note: Scroll position is not exactly at the top, current Y position:", currentScrollYInt)
	}

	return scrollYInt, nil
}

func (s *Scraper) cacheDataInRedis(url string, screenshots []string) {
	cachedData := CachedData{
		Screenshots: screenshots,
	}
	jsonData, err := json.Marshal(cachedData)
	if err != nil {
		log.Printf("Error marshaling cached data: %v", err)
		return
	}
	_, err = s.RedisClient.Set(context.Background(), url, jsonData, 24*time.Hour).Result()
	if err != nil {
		log.Printf("Error saving cached data to Redis: %v", err)
	}
}

func (s *Scraper) sendWebSocketMessage(conn *websocket.Conn, msg WebSocketMessage) {
	message, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal WebSocket message: %v", err)
		return
	}
	if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
		log.Printf("Failed to send WebSocket message: %v", err)
	}
}
