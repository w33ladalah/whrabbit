# Whrabbit

An unofficial WhatsApp API written in Go using the whatsmeow library. This project provides a REST API interface to interact with WhatsApp.

## Features

- Send text messages
- Send image messages with optional captions
- Check connection status
- QR code-based authentication
- SQLite database for session storage
- Swagger API documentation
- WebSocket support for real-time updates
- API key authentication

## Prerequisites

- Go 1.21 or higher
- SQLite3
- WhatsApp account

## Installation

1. Clone the repository:
```bash
git clone https://github.com/w33ladalah/whrabbit.git
cd whrabbit
```

2. Install dependencies:
```bash
go mod tidy
```

3. Set up environment variables:
```bash
# Development mode (using .env file)
cp .env.example .env
# Edit .env with your configuration

# Or set environment variables directly
export API_KEY=your_api_key_here
export BASE_URL=http://localhost:8080
export PORT=8080
export APP_NAME=whrabbit
export APP_VERSION=1.0.0
```

4. Build the application:
```bash
# Development build
go build -o whrabbit

# Production build with environment variables
go build -ldflags "-X github.com/w33ladalah/whrabbit/internal/config.APIKey=your_api_key_here -X github.com/w33ladalah/whrabbit/internal/config.AppVersion=1.0.0" -o whrabbit
```

## Usage

1. Run the application:
```bash
./whrabbit
```

2. Open your browser and navigate to `http://localhost:8080`

3. Scan the displayed QR code with WhatsApp on your phone to authenticate.

4. Once authenticated, the API will be available at `http://localhost:8080/api/v1`

## API Documentation

The API documentation is available at `http://localhost:8080/swagger/index.html` when the server is running.

### Authentication

All API endpoints require authentication using an API key. Include the API key in the `Authorization` header:

```
Authorization: Bearer your_api_key_here
```

### Endpoints

#### WebSocket Connection
```
GET /ws
```
Establishes a WebSocket connection to receive WhatsApp QR codes and connection status updates.

#### Send Text Message
```
POST /api/v1/messages/text
Content-Type: application/json

{
    "to": "1234567890",
    "message": "Hello, World!"
}
```

#### Send Image Message
```
POST /api/v1/messages/image
Content-Type: multipart/form-data

Form fields:
- to: Recipient's phone number (required)
- image: Image file (required)
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| API_KEY | API key for authentication | (required) |
| BASE_URL | Base URL of the application | http://localhost:8080 |
| PORT | Port to run the server on | 8080 |
| APP_NAME | Name of the application | whrabbit |
| APP_VERSION | Version of the application | dev |

## Development

### Project Structure
```
.
├── docs/               # Swagger documentation
├── internal/
│   ├── api/           # API handlers and middleware
│   ├── config/        # Configuration management
│   └── whatsapp/      # WhatsApp client implementation
├── static/            # Static files (HTML, CSS)
├── .env               # Environment variables (development)
├── main.go           # Application entry point
└── README.md         # This file
```

### Running Tests
```bash
go test ./...
```

### Generating Swagger Documentation
```bash
swag init -g main.go
```

## License

MIT

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Acknowledgments

- [whatsmeow](https://github.com/tulir/whatsmeow) - WhatsApp Web client library
- [Gin](https://github.com/gin-gonic/gin) - Web framework
- [Swagger](https://swagger.io/) - API documentation
