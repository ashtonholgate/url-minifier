# URL Minifier

A production-grade URL shortening service built with modern cloud-native technologies. This project serves as a learning platform for implementing DevOps practices, cloud deployment, and microservices architecture.

## Project Overview

This URL minifier service allows users to:
- Convert long URLs into short, manageable links
- Access analytics for link usage
- Manage their shortened URLs through a simple API
- Access the service through both REST API and web interface

## Technology Stack

### Core Technologies
- **Backend**: Go (with Gorilla/mux router)
- **Database**: MongoDB for scalable NoSQL storage
- **Cache**: Redis for high-performance URL lookups
- **Frontend**: React with TypeScript
- **Feature Flags**: Unleash for feature management

### Infrastructure & DevOps
- **Cloud Platform**: Google Cloud Platform (GCP)
  - Cloud Run for containerized services
  - Cloud SQL for managed database
  - Cloud Memorystore for Redis
  - Cloud Load Balancing
  - Cloud Monitoring and Logging

- **Infrastructure as Code**
  - Terraform for infrastructure provisioning
  - Helm charts for Kubernetes deployments

- **CI/CD**
  - GitHub Actions for automated testing and deployment
  - Docker for containerization
  - Container Registry on GCP

### Monitoring & Observability
- **Logging**: Cloud Logging
- **Metrics**: Prometheus
- **Visualization**: Grafana
- **Tracing**: OpenTelemetry

## Project Goals

1. **Learning Objectives**
   - Implement a production-grade microservices architecture
   - Master CI/CD practices using GitHub Actions
   - Gain hands-on experience with Infrastructure as Code using Terraform
   - Learn containerization and orchestration with Docker
   - Understand cloud deployment and scaling on GCP
   - Implement comprehensive logging and monitoring

2. **Technical Goals**
   - Achieve high availability and fault tolerance
   - Implement efficient caching strategies
   - Handle URL collisions and duplicates
   - Provide real-time analytics
   - Ensure security best practices

## Project Structure

```
├── services/           # All microservices
│   ├── shortener/     # URL shortening service
│   ├── analytics/     # Analytics service
│   ├── auth/          # Authentication service
│   ├── unleash/       # Feature flag service
│   │   ├── config/    # Unleash configuration
│   │   └── db/        # PostgreSQL schema for Unleash
│   └── ui/            # Frontend React application
│       ├── src/       # Main application code
│       ├── types/     # TypeScript interfaces and types
│       ├── utils/     # Frontend utilities
│       └── components/# Reusable React components
├── pkg/               # Shared Go packages
│   ├── common/        # Common Go utilities and helpers
│   ├── middleware/    # Shared Go middleware
│   ├── models/        # Go data models
│   ├── client/        # Go HTTP clients for inter-service communication
│   └── feature/       # Feature flag client wrapper
├── api/               # OpenAPI specifications
│   ├── shortener.yaml # URL shortener service API spec
│   ├── analytics.yaml # Analytics service API spec
│   └── auth.yaml      # Auth service API spec
├── deploy/            # Deployment configurations
│   ├── terraform/     # Infrastructure as code
│   ├── k8s/          # Kubernetes manifests
│   │   └── unleash/  # Unleash Kubernetes configs
│   └── docker/        # Dockerfiles for each service
├── .github/           # GitHub Actions workflows
└── docs/             # Documentation
    ├── architecture/ # Architecture diagrams and decisions
    ├── api/          # API documentation
    └── runbooks/     # Operational runbooks
```

## Code Organization

### Go Shared Code (`pkg/`)
The `pkg/` directory contains shared Go packages used across backend microservices:
- `common/`: Shared utilities like logging, config management, error handling
- `middleware/`: Common HTTP middleware (auth, logging, tracing)
- `models/`: Go structs for database models and DTOs
- `client/`: Type-safe Go clients for inter-service communication
- `feature/`: Feature flag client wrapper

### Frontend Code (`services/ui/`)
The UI service contains all frontend-related code:
- `src/`: Main application React code
- `types/`: TypeScript interfaces generated from OpenAPI specs
- `utils/`: Frontend utilities (date formatting, validation, etc.)
- `components/`: React components

### API Specifications (`api/`)
The `api/` directory serves as the single source of truth for all service interfaces:
- Contains OpenAPI/Swagger specifications for each service
- Used to generate:
  - TypeScript types and API clients for the frontend
  - Go server stubs and client libraries
  - API documentation
- Changes to these specs trigger automated code generation in both frontend and backend

## Microservices Architecture

The project is structured as a monorepo containing multiple microservices:

1. **URL Shortener Service** (`services/shortener/`)
   - Core URL shortening logic
   - URL redirection
   - URL management API

2. **Analytics Service** (`services/analytics/`)
   - Click tracking
   - Usage statistics
   - Reporting API

3. **Auth Service** (`services/auth/`)
   - User authentication
   - API key management
   - Access control

4. **UI Service** (`services/ui/`)
   - React frontend application
   - User interface for URL management
   - Analytics dashboard

5. **Feature Flag Service** (`services/unleash/`)
   - Feature flag management
   - API version transitions
   - A/B testing
   - Environment-specific configurations

Each service:
- Is independently deployable
- Has its own database connection
- Can be scaled independently
- Communicates via well-defined APIs

## Feature Flag Management

We use Unleash as our feature flag management system to handle:
- API version transitions
- Feature rollouts
- A/B testing
- Environment-specific configurations

### Integration Examples

#### Go Services
```go
// pkg/feature/client.go
package feature

import "github.com/Unleash/unleash-client-go/v4"

type Client struct {
    unleash *unleash.Client
}

func NewClient(serviceName string) (*Client, error) {
    client, err := unleash.Initialize(
        unleash.WithUrl("http://unleash-service/api/"),
        unleash.WithAppName(serviceName),
    )
    return &Client{unleash: client}, err
}

func (c *Client) IsApiV2Enabled() bool {
    return c.unleash.IsEnabled("use-new-api-v2")
}
```

#### React Frontend
```typescript
// services/ui/src/utils/features.ts
import { FlagProvider, useFlag } from '@unleash/proxy-client-react';

export const FeatureProvider: React.FC = ({ children }) => (
  <FlagProvider
    config={{
      url: '/api/proxy/unleash',
      clientKey: 'your-client-key',
      refreshInterval: 15,
      appName: 'url-minifier-ui'
    }}
  >
    {children}
  </FlagProvider>
);

export const useApiVersion = () => {
  return useFlag('use-new-api-v2') ? 'v2' : 'v1';
};
```

### Feature Flag Workflow

1. **API Changes**
   - Create new API version in OpenAPI specs
   - Add feature flag in Unleash UI
   - Generate new types with version support
   - Services implement new version behind flag

2. **Deployment**
   - Deploy updated services with both versions
   - Use Unleash UI to control rollout
   - Monitor for issues
   - Gradually increase flag adoption

3. **Cleanup**
   - Once all services use new version
   - Remove old version support
   - Archive feature flag

### Local Development

For local development, Unleash is included in the Docker Compose setup:
```yaml
# deploy/docker/docker-compose.yml
unleash:
  image: unleashorg/unleash-server
  ports:
    - "4242:4242"
  environment:
    DATABASE_URL: "postgres://postgres:postgres@postgres:5432/unleash"
    DATABASE_SSL: "false"
```

Access the Unleash UI at `http://localhost:4242` during local development.

## Development

Instructions for local development setup will be added as the project progresses.

## Deployment

Deployment instructions and pipeline documentation will be added as the CI/CD infrastructure is implemented.

## Contributing

This is a personal learning project, but suggestions and discussions are welcome through issues and discussions.

## License

MIT License - See LICENSE file for details
