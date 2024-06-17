package scraper

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
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
		if e, ok := ev.(*page.EventLifecycleEvent); ok {
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

func (s *Scraper) navigateAndSetup(url string) (context.Context, context.CancelFunc, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-notifications", true),
		chromedp.Flag("block-new-web-contents", true),
		chromedp.Flag("disable-popup-blocking", false))
	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)

	ctx, cancel := chromedp.NewContext(allocCtx)
	ctx, innerCancel := context.WithTimeout(ctx, 300*time.Second) // Increased timeout to 300 seconds

	retries := 3
	for i := 0; i < retries; i++ {
		if err := chromedp.Run(ctx, enableLifeCycleEvents(), navigateAndWaitFor(url, "networkIdle"), chromedp.Sleep(1000*time.Millisecond), chromedp.KeyEvent(kb.Escape)); err != nil {
			log.Println("Failed to navigate to:", url, "Attempt:", i+1, "Error:", err)
			time.Sleep(200 * time.Millisecond)
			continue
		}
		log.Println("Navigation completed to:", url)

		return ctx, func() {
			innerCancel()
			cancel()
		}, nil
	}
	innerCancel()
	cancel()
	return nil, nil, fmt.Errorf("failed to navigate to %s after %d attempts", url, retries)
}
