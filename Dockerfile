FROM golang:1.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o proxy-server

FROM alpine:3.16

WORKDIR /app

COPY --from=builder /app/proxy-server /app/proxy-server

EXPOSE 8080

CMD ["/app/proxy-server"]
