# Use the official Go image as the base image
FROM golang:1.22

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files to the working directory
COPY go.mod go.sum ./

# Download and cache the Go module dependencies
RUN go mod download

# Copy the games service source code to the working directory
COPY . .

# Build the games service executable
RUN go build -o auth-service main.go

# Expose the port on which the games service will run
EXPOSE 8080

# Set the entry point for the container
CMD ["./auth-service"]