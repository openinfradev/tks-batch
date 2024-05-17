FROM --platform=linux/amd64 docker.io/library/golang:1.21 AS builder

RUN mkdir -p /app
WORKDIR /app

COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/server ./cmd/server

EXPOSE 9110

ENTRYPOINT ["/app/server"]
CMD ["-port", "9110"]
