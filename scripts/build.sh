#!/bin/bash

# Read .env file
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Build with ldflags
go build -ldflags "\
    -X 'github.com/hendrowibowo/whrabbit/internal/config.APIKey=${API_KEY}' \
    -X 'github.com/hendrowibowo/whrabbit/internal/config.BaseURL=${BASE_URL}' \
    -X 'github.com/hendrowibowo/whrabbit/internal/config.ServerPort=${PORT}'" \
    -o whrabbit
