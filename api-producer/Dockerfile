# Build Stage
FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /api-producer .

# Final Stage
FROM alpine:latest
COPY --from=builder /api-producer /api-producer
EXPOSE 8080
CMD ["/api-producer"]
