FROM golang:1.25-alpine AS dev
RUN go install github.com/air-verse/air@latest
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
CMD ["air"]

FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /evently ./cmd/api

FROM alpine:3.20 AS final
RUN apk --no-cache add ca-certificates
COPY --from=builder /evently /evently
EXPOSE 8080
ENTRYPOINT ["/evently"]
