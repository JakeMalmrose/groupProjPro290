# Use the official Go image as the base image
FROM golang:1.19

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files to the working directory
COPY go.mod go.sum ./

# Download and cache the Go module dependencies
RUN go mod download

# Copy the application source code to the working directory
COPY cmd/ ./cmd/
COPY static/ ./static/
COPY templates/ ./templates/

# Build the Go application
RUN go build -o main ./cmd/server

# Expose the port on which the application will run (adjust if necessary)
EXPOSE 8080

# Set the entry point for the container
CMD ["./main"]