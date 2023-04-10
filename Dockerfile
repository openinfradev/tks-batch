FROM --platform=linux/amd64 docker.io/library/golang:1.18-buster AS builder

RUN mkdir -p /build
WORKDIR /build

COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/server ./cmd/server

EXPOSE 9110

ENTRYPOINT ["/build/bin/server"]
CMD ["-port", "9110"]
