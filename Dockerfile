FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/debravinh .

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/debravinh /app/debravinh

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
  CMD wget -qO- http://localhost:8080/ > /dev/null || exit 1

USER nobody

CMD ["/app/debravinh"]
