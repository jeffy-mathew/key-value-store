#!/usr/bin/env sh

# The code within this file will run every time you hit "RUN". 
#
# If you need a special setup to run your tests, modify this file.
#
# Alternatively you can use the terminal window below, to run any commands, like:
# go mod download
# go test ./...
# go run cmd/store/main.go

echo "==> Running the tests..."
go mod download 
go test ./...
