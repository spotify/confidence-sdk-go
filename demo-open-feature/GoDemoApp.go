package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/open-feature/go-sdk/openfeature"
	c "github.com/spotify/confidence-sdk-go/confidence"
	p "github.com/spotify/confidence-sdk-go/provider"
)

func main() {
	clientSecret := "API_KEY"
	fmt.Println("Fetching the flags...")

	confidence := c.NewConfidenceBuilder().SetAPIConfig(c.APIConfig{APIKey: clientSecret}).Build()

	confidence.PutContext("hello", "world")

	provider := p.NewFlagProvider(confidence)

	openfeature.SetProvider(provider)
	client := openfeature.NewClient("testApp")

	attributes := make(map[string]interface{})
	targetingKey := uuid.New().String()

	fmt.Println(" Random UUID -> " + targetingKey)

	of := openfeature.NewEvaluationContext(targetingKey, attributes)

	colorValue, _ := client.StringValue(context.Background(), "hawkflag.color", "defaultValue", of)
	messageValue, _ := client.StringValue(context.Background(), "hawkflag.message", "defaultValue", of)

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

	wg := confidence.Track(context.Background(), "navigate", map[string]interface{}{"test": "value"})
	wg.Wait()
}
