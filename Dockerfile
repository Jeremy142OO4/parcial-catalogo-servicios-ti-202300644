FROM golang:1.24 AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/catalogo-server ./backend/cmd/server \
    && CGO_ENABLED=0 GOOS=linux go build -o /out/catalogo-seed ./backend/cmd/seed

FROM alpine:3.21

RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/catalogo-server ./catalogo-server
COPY --from=builder /out/catalogo-seed ./catalogo-seed
COPY backend/migrations ./backend/migrations
EXPOSE 8080
CMD ["./catalogo-server"]
