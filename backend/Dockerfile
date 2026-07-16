FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o main main.go

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata postgresql-dev \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/main .
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh \
    && mkdir -p /app/media/avatar \
    && chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

ENTRYPOINT ["./entrypoint.sh"]
