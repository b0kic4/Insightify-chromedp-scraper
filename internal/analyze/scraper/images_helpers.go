package scraper

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	googleStorage "cloud.google.com/go/storage"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func (s *Scraper) captureScreenshots(conn *websocket.Conn, ctx context.Context, lastScrollY int) []string {
	var screenshots []string
	currentScrollY := 0
	scrollIncrement := 750

	s.sendWebSocketMessage(conn, WebSocketMessage{Type: "status", Content: "Content Capturing has started"})

	totalHeight := lastScrollY
	steps := totalHeight / scrollIncrement
	if totalHeight%scrollIncrement != 0 {
		steps++
	}

	// Scroll, capture screenshot and extract code
	for i := 0; i < steps; i++ {
		var screenshot []byte
		err := s.scrollAndCapture(ctx, &screenshot, &currentScrollY, scrollIncrement)
		if err != nil {
			fmt.Println("Error during scrolling and capture:", err)
			s.sendWebSocketMessage(conn, WebSocketMessage{Type: "error", Content: "Error during scrolling and capture"})
			break
		}

		if currentScrollY >= lastScrollY {
			// End of page reached, no need to capture screenshots anymore
			break
		}

		screenshotURL := s.uploadScreenshot(ctx, string(screenshot), len(screenshots))
		screenshots = append(screenshots, screenshotURL)
		fmt.Println("Screenshot captured and uploaded:", screenshotURL)
		fmt.Println("currentScrollY: ", currentScrollY)

		progress := float64(i+1) / float64(steps) * 100
		// progressMessage := fmt.Sprintf("Capturing screenshots: %.0f%% completed", progress)
		s.sendWebSocketMessage(conn, WebSocketMessage{Type: "progress", Content: progress})
	}

	s.sendWebSocketMessage(conn, WebSocketMessage{Type: "status", Content: "Analysis completed"})
	return screenshots
}

// NOTE: Extract code from extracted code based on the screenshot

// ExtractCode extracts the HTML code of the page
func (s *Scraper) extractCode(ctx context.Context) string {
	var fullHTML string
	err := chromedp.Run(ctx,
		chromedp.Evaluate(`(() => {
			const uniqueVisibleElements = new Set();
			const visibleElements = Array.from(document.body.querySelectorAll('*')).filter(el => {
				const style = window.getComputedStyle(el);
				if (el.tagName.toLowerCase() !== 'script') {  // Exclude script elements
					let parent = el.parentElement;
					while (parent) {
						if (uniqueVisibleElements.has(parent)) {
							return false;
						}
						parent = parent.parentElement;
					}
					uniqueVisibleElements.add(el);
					return true;
				}
				return false;
			});
			return Array.from(uniqueVisibleElements).map(el => el.outerHTML).join('\n');
		})()`, &fullHTML),
	)
	if err != nil {
		fmt.Println("Error during HTML extraction:", err)
		return ""
	}

	return fullHTML
}

// NOTE: Based on the lenght of the html code and the
// count of the screenshots I need to implement
// functionality for dividing the html into segments

// ScrollAndCapture performs incremental scrolls and captures screenshots
func (s *Scraper) scrollAndCapture(ctx context.Context, screenshot *[]byte, currentScrollY *int, scrollIncrement int) error {
	return chromedp.Run(ctx,
		chromedp.EmulateViewport(1920, 1080),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.CaptureScreenshot(screenshot),
		incrementalScroll(ctx, scrollIncrement),
		chromedp.Evaluate(`Math.round(window.scrollY)`, currentScrollY),
	)
}

// UploadScreenshot uploads the screenshot to Firebase Storage and returns the URL
func (s *Scraper) uploadScreenshot(ctx context.Context, screenshotData string, index int) string {
	dateFolder := time.Now().Format("2006-01-02")
	uuid, err := uuid.NewRandom()
	if err != nil {
		log.Printf("Failed to generate UUID: %v", err)
		return ""
	}
	fileName := fmt.Sprintf("%s/screenshot-%d-%s.webp", dateFolder, index, uuid)

	bucket, err := s.FirebaseStorage.Bucket(os.Getenv("FIREBASE_STORAGE_BUCKET"))
	if err != nil {
		log.Printf("Failed to get Firebase Storage bucket: %v", err)
		return ""
	}

	wc := bucket.Object(fileName).NewWriter(ctx)
	wc.ContentType = "image/webp"
	if _, err := wc.Write([]byte(screenshotData)); err != nil {
		log.Printf("Failed to write screenshot to Cloud Storage: %v", err)
		wc.Close() // Ensure the writer is closed even on failure
		return ""
	}
	if err := wc.Close(); err != nil {
		log.Printf("Failed to close Cloud Storage writer: %v", err)
		return ""
	}

	acl := bucket.Object(fileName).ACL()
	if err := acl.Set(ctx, googleStorage.AllUsers, googleStorage.RoleReader); err != nil {
		log.Printf("Failed to set public read ACL on screenshot: %v", err)
		return ""
	}

	return "https://storage.googleapis.com/" + os.Getenv("FIREBASE_STORAGE_BUCKET") + "/" + fileName
}
