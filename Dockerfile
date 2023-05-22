FROM golang:1.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/debravinh

FROM alpine:3.16

WORKDIR /app

COPY --from=builder /app/debravinh /app/debravinh
COPY web /app/web


EXPOSE 8080

CMD ["/app/debravinh"]
