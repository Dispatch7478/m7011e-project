# ---- Build stage ----
FROM golang:1.25 AS build

# Set working directory inside container
WORKDIR /app

# Copy go module files and download deps first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o user-service

# ---- Run stage ----
FROM alpine:3.20

# Create non-root user (good practice)
RUN adduser -D appuser
USER appuser

WORKDIR /app

# Copy the built binary from the previous stage
COPY --from=build /app/user-service .

# Expose the port your service listens on
EXPOSE 80



# Run the binary
ENTRYPOINT ["./user-service"]
