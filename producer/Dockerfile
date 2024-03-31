# Use the official Golang image as the base image
FROM golang:1.20

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download and install Go module dependencies
RUN go mod download

# Copy the Go source code
COPY . .

# Build the Go application
RUN go build -o main

# Expose the port for the Gin server
EXPOSE 8080

# Copy the client.properties file into the container
COPY client.properties .

# Set the entry point for the container
CMD ["./main"]