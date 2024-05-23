module demo

go 1.22.2

require (
	github.com/google/uuid v1.6.0
	github.com/open-feature/go-sdk v1.10.0
	github.com/spotify/confidence-openfeature-provider-go/confidence v1.0.0
	github.com/spotify/confidence-openfeature-provider-go/provider v0.1.7
)

replace github.com/spotify/confidence-openfeature-provider-go/provider v0.1.7 => github.com/spotify/confidence-openfeature-provider-go/provider v0.0.0-20240522103449-d3e86fdf3eab

replace github.com/spotify/confidence-openfeature-provider-go/confidence v1.0.0 => github.com/spotify/confidence-openfeature-provider-go/confidence v0.0.0-20240522103449-d3e86fdf3eab

require (
	github.com/go-logr/logr v1.4.1 // indirect
	golang.org/x/exp v0.0.0-20240213143201-ec583247a57a // indirect
)
