#!/bin/bash

# Read .env file
if [ -f .env ]; then
    export "$(grep -v '^#' .env | xargs)"
fi

# Build with ldflags
go build -ldflags "\
    -X 'github.com/w33ladalah/whrabbit/internal/config.AppName=${APP_NAME:-whrabbit}' \
    -X 'github.com/w33ladalah/whrabbit/internal/config.AppVersion=${APP_VERSION:-1.0.0}' \
    -X 'github.com/w33ladalah/whrabbit/internal/config.APIKey=${API_KEY}' \
    -X 'github.com/w33ladalah/whrabbit/internal/config.BaseURL=${BASE_URL}' \
    -X 'github.com/w33ladalah/whrabbit/internal/config.ServerPort=${PORT}'" \
    -o build/whrabbit
