package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/exp/slog"

	c "github.com/spotify/confidence-sdk-go/pkg/confidence"
)

func main() {

	// Set up the logger with the debug log level
	var programLevel = new(slog.LevelVar)
	programLevel.Set(slog.LevelDebug)
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(h))

	// Read the client secret from env
	clientkey := os.Getenv("CONFIDENCE_GO_CLIENT")
	if clientkey == "" {
		fmt.Println("No client key defined")
		os.Exit(1)
	}

	confidence := c.NewConfidenceBuilder().SetAPIConfig(*c.NewAPIConfig(clientkey)).Build().WithContext(map[string]interface{}{
		"visitor_id": "test",
	})
	colorValue := confidence.GetStringFlag(context.Background(), "hawkflag.color", "defaultValue")

	otherApiConfig := c.NewAPIConfig(clientkey)
	otherApiConfig.ResolveTimeout = 5 * time.Millisecond // setting a really short timeout
	timingOutConfidence := c.NewConfidenceBuilder().SetAPIConfig(*otherApiConfig).Build().WithContext(map[string]interface{}{
		"visitor_id": "test",
	})

	fmt.Println("Fetching the flags...")
	messageValue := timingOutConfidence.GetStringFlag(context.Background(), "hawkflag.message", "defaultValue")

	colorYellow := "\033[33m"
	colorGreen := "\033[32m"
	colorRed := "\033[31m"
	colorDefault := "\033[0m"

	fmt.Println(" Color Value --> " + colorValue.Value)
	fmt.Println(" Color Reason --> " + colorValue.Reason)
	fmt.Println(" Color Variant --> " + colorValue.Variant)
	fmt.Println(" Color ErrorCode --> " + colorValue.ErrorCode)
	fmt.Println(" Color ErrorMessage --> " + colorValue.ErrorMessage)

	fmt.Println(" Message Value --> " + messageValue.Value)
	fmt.Println(" Message Reason --> " + messageValue.Reason)
	fmt.Println(" Message Variant --> " + messageValue.Variant)
	fmt.Println(" Message ErrorCode --> " + messageValue.ErrorCode)
	fmt.Println(" Message ErrorMessage --> " + messageValue.ErrorMessage)

	switch {
	case colorValue.Value == "Yellow":
		fmt.Println(colorYellow, "Message --> "+messageValue.Value)
	case colorValue.Value == "Green":
		fmt.Println(colorGreen, "Message --> "+messageValue.Value)
	default:
		fmt.Println(colorRed, "Message --> "+messageValue.Value)
	}
	fmt.Print(colorDefault, "")
	/*
	   	wg := confidence.Track(context.Background(), "checkout-complete", map[string]interface{}{
	   		"orderId": 1234,
	   		"total":   100.0,
	   		"items":   []string{"item1", "item2"},
	   	})

	   wg.Wait()
	   fmt.Println("Event sent")
	*/
}
