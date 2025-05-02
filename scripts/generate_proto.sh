#!/bin/bash

# Install protoc
brew install protobuf

# Install protoc-gen-go
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.33.0

# Add Go binary path to PATH
export PATH="$PATH:$(go env GOPATH)/bin"

# Generate protobuf code
protoc --go_out=. --go_opt=paths=source_relative pkg/confidence/telemetry.proto 