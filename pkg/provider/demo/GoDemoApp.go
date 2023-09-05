package main

import (
	"context"
	"fmt"
	"github.com/open-feature/go-sdk/pkg/openfeature"
	confidence "github.com/spotify/confidence-openfeature-provider-go/pkg/provider"
)

func main() {
	clientSecret := "CLIENT_SECRET"
	fmt.Println("Fetching the hawk flag...")

	provider, err := confidence.NewFlagProvider(confidence.APIConfig{APIKey: clientSecret, Region: confidence.APIRegionEU})

	if err != nil {
		// handle error
	}

	openfeature.SetProvider(provider)
	client := openfeature.NewClient("testApp")

	attributes := make(map[string]interface{})
	attributes["targeting_key"] = "21339198230981240971409127491"

	colorValue, _ := client.StringValue(context.Background(), "hawkflag.color", "defaultValue",
		openfeature.NewEvaluationContext("", attributes))

	messageValue, _ := client.StringValue(context.Background(), "hawkflag.message", "defaultValue",
		openfeature.NewEvaluationContext("", attributes))

	colorYellow := "\033[33m"
	colorRed := "\033[31m"

	fmt.Println("Color --> " + colorValue)

	if colorValue == "Yellow" {
		fmt.Println("Message -->"+colorYellow, messageValue)
	} else {
		fmt.Println("Message -->"+colorRed, messageValue)
	}
}
