module github.com/spotify/confidence-openfeature-provider-go/provider

go 1.22.2

replace github.com/spotify/confidence-openfeature-provider-go/confidence v1.0.0 => ../confidence

require (
	github.com/open-feature/go-sdk v1.10.0
	github.com/spotify/confidence-openfeature-provider-go/confidence v1.0.0
	github.com/stretchr/testify v1.9.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20240213143201-ec583247a57a // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)