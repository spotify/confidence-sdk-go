package main

import (
	"context"
	"fmt"
	c "github.com/spotify/confidence-sdk-go/pkg/confidence"
)

func main() {
	fmt.Println("Fetching the flags...")

	confidence := c.NewConfidenceBuilder().SetAPIConfig(c.APIConfig{APIKey: "API_KEY"}).Build()
	targetingKey := "Random_targeting_key"
	confidence.PutContext("targeting_key", targetingKey)

	colorValue := confidence.GetStringFlag(context.Background(), "hawkflag.color", "defaultValue").Value
	messageValue := confidence.GetStringFlag(context.Background(), "hawkflag.message", "defaultValue").Value

	colorYellow := "\033[33m"
	colorGreen := "\033[32m"
	colorRed := "\033[31m"

	fmt.Println(" Color --> " + colorValue)

	switch {
	case colorValue == "Yellow":
		fmt.Println(colorYellow, "Message --> "+messageValue)
	case colorValue == "Green":
		fmt.Println(colorGreen, "Message --> "+messageValue)
	default:
		fmt.Println(colorRed, "Message --> "+messageValue)
	}

	wg := confidence.Track(context.Background(), "page-viewed", map[string]interface{}{})
	wg.Wait()
	fmt.Println("Event sent")

}
