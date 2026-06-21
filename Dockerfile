FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/service

FROM alpine:3.20

RUN apk add --no-cache wget ca-certificates tzdata

WORKDIR /app
COPY --from=builder /build/server .
COPY migrations/ migrations/

EXPOSE 8080 8081

CMD ["./server"]
