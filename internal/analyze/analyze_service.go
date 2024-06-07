package analyze

import (
	"Insightify-backend/internal/analyze/scraper"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Command struct {
	URL      string `json:"url"`
	Market   string `json:"market"`
	Audience string `json:"audience"`
	Insights string `json:"insights"`
	UserID   string `json:"userID"`
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		var cmd Command
		if err := json.Unmarshal(message, &cmd); err != nil {
			log.Printf("Error unmarshaling command: %v", err)
			continue
		}
		ctx := r.Context()
		scraperInstance := scraper.NewScraper(ctx)
		screenshotURLs := scraperInstance.CaptureAndUpload(cmd.URL, cmd.UserID, cmd.Market, cmd.Audience, cmd.Insights, conn)

		// Send results back to the client
		response, err := json.Marshal(screenshotURLs)
		if err != nil {
			log.Printf("Error marshaling response: %v", err)
			continue
		}

		if err := conn.WriteMessage(websocket.TextMessage, response); err != nil {
			log.Printf("Error sending response: %v", err)
			break
		}
	}
}
