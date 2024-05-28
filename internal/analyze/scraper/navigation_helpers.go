package scraper

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/gorilla/websocket"
)

func enableLifeCycleEvents() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		if err := page.Enable().Do(ctx); err != nil {
			return err
		}
		return page.SetLifecycleEventsEnabled(true).Do(ctx)
	}
}

func navigateAndWaitFor(url string, eventName string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		_, _, _, err := page.Navigate(url).Do(ctx)
		if err != nil {
			log.Println("Error in navigateAndWaitFor: ", err)
			return err
		}
		return waitFor(ctx, eventName)
	}
}

func waitFor(ctx context.Context, eventName string) error {
	ch := make(chan struct{})
	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	chromedp.ListenTarget(cctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *page.EventLifecycleEvent:
			if e.Name == eventName {
				cancel()
				close(ch)
			}
		}
	})
	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func incrementalScroll(ctx context.Context, scrollIncrement int) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		script := fmt.Sprintf("window.scrollBy(0, %d);", scrollIncrement)
		return chromedp.Evaluate(script, nil).Do(ctx)
	}
}

func (s *Scraper) navigateAndSetup(url string, conn *websocket.Conn) (context.Context, context.CancelFunc, error) {
	fmt.Println("in the navigate and setup")
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"),
	)
	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	// defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	ctx, innerCancel := context.WithTimeout(ctx, 300*time.Second) // Increased timeout to 300 seconds

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if req, ok := ev.(*network.EventRequestWillBeSent); ok {
			log.Printf("Request URL: %s\n", req.Request.URL)
		}
	})

	s.sendWebSocketMessage(conn, WebSocketMessage{Type: "status", Content: "Navigating to: " + url})

	retries := 3
	for i := 0; i < retries; i++ {
		fmt.Println("in the for loop of navigate and setup i: ", i)
		if err := chromedp.Run(ctx, enableLifeCycleEvents(), navigateAndWaitFor(url, "networkIdle")); err != nil {
			log.Println("Failed to navigate to:", url, "Attempt:", i+1, "Error:", err)
			time.Sleep(2 * time.Second)
			continue
		}
		log.Println("Navigation completed to:", url)
		return ctx, func() {
			innerCancel()
			cancel()
		}, nil
	}
	return nil, nil, fmt.Errorf("failed to navigate to %s after %d attempts", url, retries)
}
