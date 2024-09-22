#!/bin/bash

# Step 1: Start Docker containers
echo "Starting Docker containers..."
docker-compose up -d

# Step 2: Wait for containers to be ready (adjust as necessary)
echo "Waiting for services to be ready..."
sleep 15

# Step 3: Build the Go application
echo "Building the application..."
make build

# Step 4: Run the Go application
echo "Running the application..."
make run

# Step 5: Optional - Create 10 test users
echo "Creating test users..."
for i in {1..10}; do
  curl -X POST http://localhost:8080/register -d '{"username":"user'$i'", "email":"user'$i'@test.com", "password":"password'$i'"}' -H "Content-Type: application/json"
done

echo "Setup, build, and run completed."
