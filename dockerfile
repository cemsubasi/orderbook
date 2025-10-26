ARG TARGETOS=linux
ARG TARGETARCH=amd64

FROM golang:1.24-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o orderbook ./cmd/main.go

RUN apk add --no-cache curl tar
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz \
    | tar xz -C /usr/local/bin

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/orderbook .
COPY --from=build /app/internal/db/migrations ./internal/db/migrations
COPY apply_migrations.sh ./apply_migrations.sh
COPY --from=build /usr/local/bin/migrate /usr/local/bin/migrate
RUN chmod +x ./apply_migrations.sh

EXPOSE 8080
ENTRYPOINT ["./orderbook"]