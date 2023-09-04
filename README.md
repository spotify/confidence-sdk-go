# Confidence OpenFeature Go Flag Provider

This repo contains the OpenFeature Go flag provider for Confidence.

## OpenFeature

Before starting to use the provider, it can be helpful to read through the general [OpenFeature docs](https://docs.openfeature.dev/)
and get familiar with the concepts. 

## Adding the dependency

Until the module is available in a public repository, you can add it to your project by 
adding something like this to your `go.mod`. 

```
replace github.com/spotify/openfeature-go/pkg/provider => ../openfeature-go/pkg/provider

require (
	github.com/spotify/openfeature-go/pkg/provider v0.0.1
	github.com/open-feature/go-sdk v1.1.0
)
```

### Creating and using the flag provider

Below is an example for how to create a OpenFeature client using the Confidence flag provider, and then resolve
a flag with a boolean attribute. The provider is configured with an api key and a region, which will determine
where it will send the resolving requests. 

The flag will be applied immediately, meaning that Confidence will count the targeted user as having received the treatment. 

You can retrieve attributes on the flag variant using property dot notation, meaning `test-flag.boolean-key` will retrieve
the attribute `boolean-key` on the flag `test-flag`. 

You can also use only the flag name `test-flag` and retrieve all values as a map with `client.ObjectValue()`. 

The flag's schema is validated against the requested data type, and if it doesn't match it will fall back to the default value. 

```go
import (
    "github.com/open-feature/go-sdk/pkg/openfeature"
    confidence "github.com/spotify/openfeature-go/pkg/provider"
)

provider, err := confidence.NewFlagProvider(confidence.APIConfig{APIKey: "apiKey", Region: confidence.APIRegionEU})

if err != nil {
    // handle error	
}

openfeature.SetProvider(provider)
client := openfeature.NewClient("testApp")
	
attributes := make(map[string]interface{})
attributes["country"] = "SE"
attributes["plan"] = "premium"
attributes["user_id"] = "dennis"

boolValue, error := client.BooleanValue(context.Background(), "test-flag.boolean-key", false, 
	openfeature.NewEvaluationContext("", attributes))
```
