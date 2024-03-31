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

# Copy the client.properties file into the container
COPY client.properties .
COPY copper-stacker-411515-a77e011c07dd.json .

# Copy the service account key file into the container
COPY copper-stacker-411515-a77e011c07dd.json /app/copper-stacker-411515-a77e011c07dd.json

# Set the GOOGLE_APPLICATION_CREDENTIALS environment variable
ENV GOOGLE_APPLICATION_CREDENTIALS="/app/copper-stacker-411515-a77e011c07dd.json"

# Set the entry point for the container
CMD ["./main"]