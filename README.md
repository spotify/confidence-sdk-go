# Confidence Go SDK

This repo contains the [Confidence](https://confidence.spotify.com/) Go SDK.

## Adding the dependency
<!---x-release-please-start-version-->
```
require (
	github.com/spotify/confidence-sdk-go v0.4.4
)
```
<!---x-release-please-end-->

## Creating and using the SDK

Below is an example for how to create an instance of the Confidence SDK, and then resolve a flag with a boolean attribute.

The SDK is configured via `SetAPIConfig(...)` and `*c.NewAPIConfig(...)`, with which you can set the api key for authentication.
Optionally, a custom resolve API url can be configured if, for example, the resolver service is running on a locally deployed side-car (`NewAPIConfigWithUrl(...)`).


You can retrieve properties on the flag variant using property dot notation, meaning `test-flag.boolean-key` will retrieve the attribute `boolean-key` on the flag `test-flag`. 

You can also use only the flag name `test-flag` and retrieve all values as a map with `GetObjectFlag()`. 

The flag's schema is validated against the requested data type, and if it doesn't match it will fall back to the default value. 

```go
import (
    c "github.com/spotify/confidence-sdk-go/pkg/confidence"
)

confidenceSdk := c.NewConfidenceBuilder().SetAPIConfig(*c.NewAPIConfig("clientSecret")).Build()

confidence.PutContext("targeting_key", "Random_targeting_key")
flagValue := confidence.GetBoolFlag(context.Background(), "test-flag.boolean-key", false).Value
// we can also pull flag values using a Confidence instance with extra context
confidence.WithContext(map[string]interface{}{
    "Something": 343,
}).GetBoolFlag(context.Background(), "test-flag.boolean-key", false).Value
```

The flag will be applied immediately, meaning that Confidence will count the targeted user as having received the treatment once they have have been evaluated. 

### Tracking

Confidence support event tracking through the SDK. The `Track()` function accepts an en event name and a map of arbitrary data connected to the event.
The current context will also be appended to the event data.

```go
wg := confidence.Track(context.Background(), "checkout-complete", map[string]interface{}{
    "orderId": 1234,
    "total":   100.0,
    "items":   []string{"item1", "item2"},
})
wg.Wait()
```

### Telemetry

The SDK includes telemetry functionality that helps monitor SDK performance and usage. By default, telemetry is enabled and collects metrics (anonymously) such as resolve latency and request status. This data is used by the Confidence team, and in certain cases it is also exposed to the SDK adopters. You can disable telemetry by setting `DisableTelemetry: true` in the `APIConfig`:

```go
config := c.NewAPIConfig("clientSecret")
config.DisableTelemetry = true
confidenceSdk := c.NewConfidenceBuilder().SetAPIConfig(*config).Build()
```

## Logging

Unless specifically configured using the `ConfidenceBuilder` `setLogger()` function; Confidence uses the default instance of [slog](https://pkg.go.dev/log/slog) for logging valuable information during runtime.
When getting started with Confidence, we suggest you configure [slog](https://pkg.go.dev/log/slog) to emit debug level information:
```go
// Set up the logger with the debug log level
var programLevel = new(slog.LevelVar)
programLevel.Set(slog.LevelDebug)
h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: programLevel})
slog.SetDefault(slog.New(h))
``` 

## Demo app

To run the demo app, replace the `CLIENT_SECRET` with client secret setup in the 
[Confidence](https://confidence.spotify.com/) console, the flags with existing flags and execute 
the app with `cd demo && go run GoDemoApp.go`.

## Confidence OpenFeature Go Provider

The SDK can be combined with the [OpenFeature Go SDK](https://github.com/open-feature/go-sdk), the repo also contains an OpenFeature Provider. Before starting to use the provider, it can be helpful to read through the general [OpenFeature docs](https://docs.openfeature.dev/)
and get familiar with the concepts. 
It's also important to add the underlying OpenFeature SDK dependency:
```
require (
	github.com/open-feature/go-sdk v1.7.0
)
```

### Creating and using the flag provider

Below is an example for how to create a OpenFeature client using the Confidence flag provider, and then resolve
a flag with a boolean attribute.

The Provider constructor accepts a confidence instance: `NewFlagProvider(confidenceSdk)`, please refer to the previous sections 
of this readme for more detailed information on how to set that up.

You can retrieve attributes on the flag variant using property dot notation, meaning `test-flag.boolean-key` will retrieve
the attribute `boolean-key` on the flag `test-flag`. 

You can also use only the flag name `test-flag` and retrieve all values as a map with `client.ObjectValue()`. 

The flag's schema is validated against the requested data type, and if it doesn't match it will fall back to the default value. 

```go
import (
    o "github.com/open-feature/go-sdk/openfeature"
    c "github.com/spotify/confidence-sdk-go/pkg/confidence"
    p "github.com/spotify/confidence-sdk-go/pkg/provider"
)

confidenceSdk := c.NewConfidenceBuilder().SetAPIConfig(*c.NewAPIConfig("clientSecret")).Build()
confidenceProvider := p.NewFlagProvider(confidenceSdk)


o.SetProvider(confidenceProvider)
client := o.NewClient("testApp")
	
attributes := make(map[string]interface{})
attributes["country"] = "SE"
attributes["plan"] = "premium"
attributes["user_id"] = "user1"

boolValue, error := client.BooleanValue(context.Background(), "test-flag.boolean-key", false, 
	o.NewEvaluationContext("", attributes))
```

