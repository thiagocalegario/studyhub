FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o studyhub ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/studyhub .
COPY web/ web/
EXPOSE 8080
CMD ["./studyhub"]