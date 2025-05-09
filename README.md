# Whrabbit

An unofficial WhatsApp API written in Go using the whatsmeow library. This project provides a REST API interface to interact with WhatsApp.

## Features

- Send text messages
- Send image messages with optional captions
- Check connection status
- QR code-based authentication
- SQLite database for session storage

## Prerequisites

- Go 1.21 or higher
- SQLite3

## Installation

1. Clone the repository:
```bash
git clone https://github.com/hendrowibowo/whrabbit.git
cd whrabbit
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
go build -o whrabbit
```

## Usage

1. Run the application:
```bash
./whrabbit
```

2. On first run, the application will display a QR code in the terminal. Scan this QR code with WhatsApp on your phone to authenticate.

3. Once authenticated, the API will be available at `http://localhost:8080`

## API Endpoints

### Send Text Message
```
POST /api/send/text
Content-Type: application/json

{
    "to": "1234567890",
    "message": "Hello, World!"
}
```

### Send Image Message
```
POST /api/send/image
Content-Type: multipart/form-data

Form fields:
- to: Recipient's phone number (required)
- image: Image file (required)
- caption: Image caption (optional)

Example using curl:
curl -X POST http://localhost:8080/api/send/image \
  -F "to=1234567890" \
  -F "image=@/path/to/image.jpg" \
  -F "caption=Check out this photo!"
```

### Check Status
```
GET /api/status
```

## Response Format

All responses are in JSON format:

Success response:
```json
{
    "status": "success",
    "message": "Message sent successfully",
    "details": {
        "filename": "image.jpg",
        "size": 123456,
        "caption": "Check out this photo!"
    }
}
```

Error response:
```json
{
    "error": "Error message here"
}
```

## License

MIT

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.
