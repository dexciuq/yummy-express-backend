# Use the official Golang image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the Go source code into the container
COPY . .

# Build the Go application
RUN go build -o app ./cmd/api

# Expose the port that your application listens on
EXPOSE 8080

# Run the Go application
CMD ["./app"]
