package main

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/exp/slog"

	c "github.com/spotify/confidence-sdk-go/pkg/confidence"
)

func main() {

	// Set up the logger with the debug log level
	var programLevel = new(slog.LevelVar)
	programLevel.Set(slog.LevelDebug)
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(h))

	confidence := c.NewConfidenceBuilder().SetAPIConfig(*c.NewAPIConfig("CLIENT_SECRET")).Build()
	confidence.PutContext("targeting_key", "Random_targeting_key")
	withAddedContext := confidence.WithContext(map[string]interface{}{
		"Something": 343,
	})

	// we can pull flag values using a Confidence instance with extra context
	fmt.Println("Fetching the flags...")
	colorValue := withAddedContext.GetStringFlag(context.Background(), "hawkflag.color", "defaultValue").Value
	messageValue := withAddedContext.GetStringFlag(context.Background(), "hawkflag.message", "defaultValue").Value

	colorYellow := "\033[33m"
	colorGreen := "\033[32m"
	colorRed := "\033[31m"
	colorDefault := "\033[0m"

	fmt.Println(" Color --> " + colorValue)

	switch {
	case colorValue == "Yellow":
		fmt.Println(colorYellow, "Message --> "+messageValue)
	case colorValue == "Green":
		fmt.Println(colorGreen, "Message --> "+messageValue)
	default:
		fmt.Println(colorRed, "Message --> "+messageValue)
	}
	fmt.Print(colorDefault, "")

	wg := confidence.Track(context.Background(), "checkout-complete", map[string]interface{}{
		"orderId": 1234,
		"total":   100.0,
		"items":   []string{"item1", "item2"},
	})
	wg.Wait()
	fmt.Println("Event sent")
}
