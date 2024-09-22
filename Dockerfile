# Start from the official Go 1.22.6 image
FROM golang:1.22.6-alpine

# Set the current working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to the container
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code to the container
COPY . .

# Build the Go app
RUN go build -o golang-chat-api ./cmd/app/

# Expose the port
EXPOSE 8080

# Set environment variables (can be overridden by docker-compose.yml or environment variables)
ENV PORT=8080
ENV DB_HOST=postgres
ENV DB_PORT=5432
ENV DB_USER=postgres
ENV DB_PASS=yourpassword
ENV DB_NAME=chatdb
ENV REDIS_HOST=redis
ENV REDIS_PORT=6379

# Run the application
CMD ["./golang-chat-api"]
