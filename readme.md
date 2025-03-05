# Ephemeral Social Space Generator (ESSG)

## Overview

The Ephemeral Social Space Generator is a novel platform that automatically detects trending conversations across multiple social media platforms, creates temporary purpose-built discussion spaces in response, and gracefully dissolves these spaces when natural engagement concludes. The platform includes geo-tagging capabilities to surface both global and hyperlocal discussions relevant to users' locations.

## Core Components

- **Social Listening Engine**: Monitors multiple platforms to detect emerging trends and conversations
- **Space Management System**: Dynamically creates and manages ephemeral discussion spaces
- **Geospatial Services**: Provides location-aware features and hyperlocal relevance
- **Engagement Analysis**: Monitors activity to determine space lifecycle stages
- **Real-time Communication**: Enables seamless interactions within spaces

## Getting Started

### Prerequisites

- Go 1.19+
- PostgreSQL 14+ with PostGIS and TimescaleDB extensions
- NATS message broker
- Node.js 16+ and NPM (for the frontend)

### Installation

1. Clone the repository
   ```
   git clone https://github.com/your-org/essg.git
   cd essg
   ```

2. Install Go dependencies
   ```
   go mod download
   ```

3. Set up the database
   ```
   psql -U postgres -d postgres -c "CREATE DATABASE essg;"
   psql -U postgres -d essg -f scripts/schema.sql
   ```

4. Configure environment variables
   ```
   cp .env.example .env
   # Edit .env with your configuration
   ```

5. Build the project
   ```
   go build -o bin/essg ./cmd/api
   ```

6. Start NATS server
   ```
   docker run -p 4222:4222 -p 8222:8222 -p 6222:6222 --name nats-server nats
   ```

7. Start the application
   ```
   ./bin/essg
   ```

### Development Setup

For development, you can use the provided Docker Compose configuration:

```
docker-compose up -d
```

This will start:
- PostgreSQL with PostGIS and TimescaleDB
- NATS server
- NATS streaming server for durable message delivery

## Project Structure

```
essg/
├── cmd/                    # Application entry points
│   ├── api/                # Main API server
│   ├── worker/             # Background workers
│   └── admin/              # Admin tools
├── internal/               # Private application code
│   ├── domain/             # Core domain models
│   ├── service/            # Application services
│   ├── adapter/            # External service adapters
│   ├── config/             # Configuration management
│   └── server/             # HTTP and WebSocket servers
├── pkg/                    # Public libraries
├── web/                    # Next.js frontend application
├── scripts/                # Build and deployment scripts
├── docs/                   # Documentation
└── test/                   # Integration and E2E tests
```

## Architecture

ESSG follows a domain-driven design approach with clean architecture principles:

1. **Domain Layer**: Contains the core business logic and interfaces
2. **Service Layer**: Implements domain interfaces and orchestrates workflows
3. **Adapter Layer**: Connects to external systems like databases and APIs
4. **Infrastructure Layer**: Provides technical capabilities like HTTP servers

The system is event-driven, using NATS for real-time communication between components.

## Key Features

### Social Listening Engine

- Cross-platform trend detection
- Natural language processing for topic clustering
- Trend velocity detection
- Geo-tagging of content

### Space Management

- Dynamic template selection based on conversation type
- Lifecycle management (growing, peak, waning, dissolving)
- Automatic feature allocation
- Location-aware space creation

### Geospatial Services

- Efficient geospatial indexing
- Support for multiple levels of locality
- Privacy-preserving location handling
- Adaptive radius based on population density

### Real-time Communication

- WebSocket-based messaging
- Message enrichment with context
- Activity-based UI updates
- Geo-context for location-based discussions

## API Documentation

The API documentation is available at `/docs/api` when the server is running, or in the `docs/api` directory.

## Frontend Development

The frontend application is built with Next.js and can be found in the `web` directory:

```
cd web
npm install
npm run dev
```

## Testing

Run the test suite:

```
go test ./...
```

Run integration tests:

```
go test ./test/integration/...
```

## Deployment

ESSG is designed to be deployed on Fly.io for the backend and Vercel for the frontend.

### Backend Deployment

```
fly launch
```

### Frontend Deployment

```
cd web
vercel
```

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.