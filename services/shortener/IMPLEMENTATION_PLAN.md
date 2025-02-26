# URL Shortener Service Implementation Plan

This document outlines the implementation plan for the URL shortener service, which is responsible for converting long URLs into short, manageable links.

## Service Structure
```
services/shortener/
├── cmd/
│   └── server/           # Main application entry point
├── internal/
│   ├── api/             # HTTP handlers and routing
│   ├── config/          # Configuration management
│   ├── domain/          # Core business logic
│   ├── repository/      # Database interactions
│   └── service/         # Service layer
├── pkg/                 # Public packages if needed
├── Dockerfile          # Container configuration
├── go.mod             # Go module definition
└── README.md          # Service-specific documentation
```

## Core Components

### URL Shortening Logic (`internal/domain`)
- URL validation and sanitization
- Short URL generation using base62 encoding
- Collision detection and handling
- URL expiration management
- Custom URL support

### Data Layer (`internal/repository`)
- MongoDB integration for URL storage
- Redis integration for caching
- Schema:
  ```go
  type URL struct {
      ID          string    `bson:"_id"`
      LongURL     string    `bson:"long_url"`
      ShortCode   string    `bson:"short_code"`
      UserID      string    `bson:"user_id"`
      CreatedAt   time.Time `bson:"created_at"`
      ExpiresAt   time.Time `bson:"expires_at"`
      CustomAlias string    `bson:"custom_alias"`
  }
  ```

### API Endpoints (`internal/api`)
- `POST /api/v1/urls` - Create short URL
- `GET /api/v1/urls/:id` - Get URL details
- `GET /:shortCode` - Redirect to long URL
- `DELETE /api/v1/urls/:id` - Delete short URL
- `PUT /api/v1/urls/:id` - Update URL details

### Service Layer (`internal/service`)
- Business logic implementation
- Integration with feature flags
- Error handling and validation
- Rate limiting
- Analytics event emission

## Technical Features

### Caching Strategy
- Redis for frequently accessed URLs
- Cache-aside pattern implementation
- TTL-based cache invalidation

### Monitoring & Observability
- OpenTelemetry integration
- Prometheus metrics
- Structured logging with correlation IDs
- Health check endpoints

### Security
- Input validation and sanitization
- Rate limiting per user/IP
- URL blacklisting support
- API authentication middleware

## Implementation Steps

1. Create OpenAPI specification (`api/shortener.yaml`)
   - Define all API endpoints and data models
   - Include detailed request/response schemas
   - Document error responses and status codes
   - Add examples and descriptions

2. Set up project structure and base Go module
   - Initialize Go module
   - Create directory structure
   - Set up basic configuration

3. Create Dockerfile and Docker Compose setup
   - Dockerfile for the shortener service
   - Docker Compose file including:
     - Go service
     - MongoDB instance
     - Redis instance
     - Environment variables and networking

4. Implement core URL shortening logic
5. Set up MongoDB and Redis connections
6. Create repository layer with database operations
7. Implement service layer business logic
8. Create HTTP handlers and API endpoints
9. Add middleware (auth, logging, metrics)
10. Implement caching layer
11. Add monitoring and observability
12. Write unit and integration tests
13. Add documentation

## Dependencies
```go
// Core dependencies
github.com/gorilla/mux          // HTTP routing
go.mongodb.org/mongo-driver     // MongoDB driver
github.com/go-redis/redis/v8    // Redis client
github.com/Unleash/unleash-client-go/v4  // Feature flags

// Observability
go.opentelemetry.io/otel       // OpenTelemetry
github.com/prometheus/client_golang // Metrics

// Utils
github.com/spf13/viper         // Configuration
github.com/rs/zerolog          // Structured logging
github.com/golang-jwt/jwt      // JWT authentication
