FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
ADD --git .git
RUN go mod download

COPY . .

RUN make build

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/bin/event-receiver .
COPY config.yaml .

EXPOSE 8000
CMD ["./event-receiver"]