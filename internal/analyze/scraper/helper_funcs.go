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

// scrollToTop attempts to scroll to the top of the page by using Home key and PageUp key.
func (s *Scraper) scrollToTop(ctx context.Context, retries int) error {
	var currentScrollY float64

	for i := 0; i <= retries; i++ {
		// Use chromedp to scroll to the top of the page
		err := chromedp.Run(ctx,
			chromedp.KeyEvent(kb.Home),
			chromedp.Sleep(1500*time.Millisecond), // Increased sleep time
			chromedp.KeyEvent(kb.PageUp),
			chromedp.Evaluate(`Math.round(window.scrollY)`, &currentScrollY),
		)
		if err != nil {
			if i == retries {
				fmt.Println("Failed to scroll back to the top:", err)
				return err
			}
			continue
		}

		if int(currentScrollY) == 0 {
			return nil // Successfully scrolled to the top
		}

		fmt.Printf("Retry %d: Scroll position is not exactly at the top, current Y position: %d\n", i+1, int(currentScrollY))
	}

	return fmt.Errorf("unable to scroll to the top after %d retries", retries)
}

func (s *Scraper) determineHeight(ctx context.Context) (int, error) {
	var lastScrollY float64
	maxRetries := 5 // Maximum number of retries
	retryCount := 0
	sleepTime := 1500 * time.Millisecond

	// Function to scroll to the bottom of the page
	scrollToBottom := func() error {
		return chromedp.Run(ctx,
			chromedp.KeyEvent(kb.End),
			chromedp.Sleep(sleepTime),
			chromedp.KeyEvent(kb.PageDown),
			chromedp.Evaluate(`Math.round(window.scrollY)`, &lastScrollY),
		)
	}

	// Retry mechanism
	for retryCount < maxRetries {
		err := scrollToBottom()
		if err != nil {
			fmt.Println("Failed to scroll to the bottom:", err)
			return 0, err
		}

		if lastScrollY > 0 {
			break
		}

		fmt.Printf("Retry %d: lastScrollY is 0, retrying...\n", retryCount+1)
		retryCount++
		sleepTime += 500 * time.Millisecond // Optionally increase sleep time on each retry
	}

	if lastScrollY == 0 {
		return 0, fmt.Errorf("failed to get a valid scroll height after %d retries", maxRetries)
	}

	scrollYInt := int(lastScrollY)
	fmt.Println("last scroll: ", scrollYInt)

	retries := 5
	err := s.scrollToTop(ctx, retries)
	if err != nil {
		return 0, err
	}

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
