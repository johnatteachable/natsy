# Build Stage
FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /consumer .

# Final Stage
FROM alpine:latest
COPY --from=builder /consumer /consumer
EXPOSE 8080
CMD ["/consumer"]
