#!/bin/bash

if [ ! -f .env ]; then
  echo "Error: Unable to find .env file. Please try to run this script from the root directory."
  exit 1
fi

# Read .env file
if [ -f .env ]; then
    export "$(grep -v '^#' .env | xargs)"
fi

# Copy static directory to build dir
if [ -d "static" ]; then
    cp -r static build/
else
    echo "Warning: static directory not found"
fi

# Build with ldflags
go build -ldflags "\
    -X 'github.com/w33ladalah/whrabbit/internal/config.AppName=${APP_NAME:-whrabbit}' \
    -X 'github.com/w33ladalah/whrabbit/internal/config.AppVersion=${APP_VERSION:-1.0.0}' \
    -X 'github.com/w33ladalah/whrabbit/internal/config.APIKey=${API_KEY}' \
    -X 'github.com/w33ladalah/whrabbit/internal/config.BaseURL=${BASE_URL}' \
    -X 'github.com/w33ladalah/whrabbit/internal/config.ServerPort=${PORT}'" \
    -o build/whrabbit
