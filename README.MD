# Golang Chat API

A fully-featured chat API built with **Golang**, **GIN**, **Redis**, **PostgreSQL**, **Prometheus**, **Grafana**, and **OpenTelemetry** for observability and monitoring.

## Features
- User Authentication with JWT
- Real-time Chat with WebSockets
- PostgreSQL for persistent data storage
- Redis for caching and WebSocket connection management
- Prometheus for collecting metrics
- Grafana for visualizing metrics and logs
- OpenTelemetry for tracing and observability
- Rate Limiting Middleware

## Prerequisites
- Docker and Docker Compose
- Golang `v1.22.6`
- Make (for running database migrations)

## Environment Variables

Create a `.env` file in the project root with the following configuration:

```bash
# App configuration
PORT=8080

# PostgreSQL database configuration
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASS=yourpassword
DB_NAME=chatdb

# Redis configuration
REDIS_HOST=redis
REDIS_PORT=6379
```


## Running the Application

Clone the repository:

```bash
git clone https://github.com/joshbarros/golang-chat-api.git
cd golang-chat-api
```

Start the services with Docker Compose:

```bash
docker-compose up -d
```

The API will be available at http://localhost:8080.

Access Grafana at http://localhost:3000 with credentials admin/admin.

Prometheus metrics will be available at http://localhost:9090.

Chat API metrics are exposed at /metrics.

## Managing Database Migrations

You can use `Make` to easily manage database migrations.

1. **Apply migrations (migrate-up)**:

  ```bash
   make migrate-up
  ```

1. **Rollback migrations (migrate-down)**:
  ```bash
  make migrate-down
  ```

## API Endpoints

- **User Registration**: POST /register
  ```json
  {
    "username": "testuser",
    "password": "password123"
  }
  ```

- **User Login**: POST /login

  ```json
  {
    "username": "testuser",
    "password": "password123"
  }
  ```

## WebSocket Chat: Connect to the WebSocket:

```bash
ws://localhost:8080/ws/{roomID}
````

## Monitoring and Observability
Prometheus scrapes metrics from the chat API at /metrics.
Grafana provides visual dashboards for performance and monitoring data.

## License
This project is licensed under the MIT License.

