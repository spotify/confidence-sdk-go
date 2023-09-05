package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/open-feature/go-sdk/pkg/openfeature"
	confidence "github.com/spotify/confidence-openfeature-provider-go/pkg/provider"
)

func main() {
	clientSecret := "CLIENT_SECRET"
	fmt.Println("Fetching the flags...")

	provider, err := confidence.NewFlagProvider(confidence.APIConfig{APIKey: clientSecret, Region: confidence.APIRegionEU})

	if err != nil {
		// handle error
	}

	openfeature.SetProvider(provider)
	client := openfeature.NewClient("testApp")

	attributes := make(map[string]interface{})
	targetingKey := uuid.New().String()
	attributes["targeting_key"] = targetingKey

	fmt.Println(" Random UUID -> " + targetingKey)

	of := openfeature.NewEvaluationContext("", attributes)

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
}
