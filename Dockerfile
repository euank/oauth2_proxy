FROM golang:latest as builder

RUN mkdir -p /app
WORKDIR /app

COPY . .

RUN go mod verify && go build -o /oauth2_proxy .

FROM debian
COPY --from=builder /oauth2_proxy /oauth2_proxy
