# Stage 1: build
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o sumit .

# Stage 2: run
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/sumit .
COPY --from=builder /app/static ./static
EXPOSE 8080
CMD ["./sumit"]
