# Freemial Server Go

A WebSocket-enabled HTTP server built in Go for managing device bindings and user authentication in the Freemial ecosystem.

## Features

- **WebSocket Communication**: Real-time bidirectional communication between clients and server
- **Device Management**: API endpoints for managing device bindings and states
- **User Authentication**: Login system with token-based authentication
- **CORS Support**: Cross-origin resource sharing enabled for web applications
- **Containerized Deployment**: Docker support for easy deployment

## Project Structure

```
freemial-server-go/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── api/
│   │   ├── device.go         # Device binding management API
│   │   └── login.go          # User authentication API
│   ├── config/               # Configuration management
│   ├── logger/               # Logging utilities
│   └── websocket/
│       ├── hub.go           # WebSocket hub management
│       ├── hub_test.go      # Hub tests
│       └── server.go        # WebSocket server implementation
├── Dockerfile               # Container build configuration
├── go.mod                   # Go module definition
└── go.sum                   # Dependency checksums
```

## API Endpoints

### Authentication
- `POST /login` - User authentication endpoint
  - Returns access tokens, refresh tokens, and user credentials

### Device Management
- `GET /device/bindings` - Retrieve device bindings
  - Returns list of bound devices with their states and metadata

### WebSocket
- `GET /` - WebSocket connection endpoint
  - Establishes real-time communication channel

## Installation

### Prerequisites
- Go 1.18 or higher
- Docker (optional, for containerized deployment)

### Local Development

1. Clone the repository:
```bash
git clone https://github.com/freemial/freemial-server-go.git
cd freemial-server-go
```

2. Install dependencies:
```bash
go mod download
```

3. Run the server:
```bash
go run cmd/server/main.go
```

The server will start on port 8080 by default.

### Custom Port

You can specify a custom port using the `-addr` flag:
```bash
go run cmd/server/main.go -addr :3000
```

## Docker Deployment

1. Build the Go binary:
```bash
go build -o server cmd/server/main.go
```

2. Build the Docker image:
```bash
docker build -t freemial-server .
```

3. Run the container:
```bash
docker run -p 8080:8080 freemial-server
```

## Dependencies

- **gorilla/websocket** (v1.5.3) - WebSocket implementation
- **oapi-codegen/nullable** (v1.1.0) - Nullable types for API schemas  

## Development

### Running Tests
```bash
go test ./...
```

### Building for Production
```bash
go build -ldflags="-w -s" -o server cmd/server/main.go
```

## Configuration

The server accepts the following command-line flags:
- `-addr`: HTTP service address (default: `:8080`)

## WebSocket Channels

The server supports multiple WebSocket channels for different devices. Each channel maintains its own set of connected clients and can broadcast messages independently.

## CORS Policy

The server is configured with permissive CORS headers for development:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Headers: *`  
- `Access-Control-Allow-Methods: DELETE,GET,HEAD,OPTIONS,PUT,POST,PATCH`

For production deployment, consider restricting these to specific origins and methods.

## License

This project is part of the Freemial ecosystem. Please refer to the project's license file for usage terms.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Support

For issues and questions, please open an issue in the project repository.
