# Use an official Go image as the base image
FROM golang:1.23

# Set the working directory in the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
ENV GO111MODULE on

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY ./pkg .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/azure-nodepool-metrics

# Expose the port the app runs on
EXPOSE 8002

# Run the app
CMD ["./azure-nodepool-metrics"]
