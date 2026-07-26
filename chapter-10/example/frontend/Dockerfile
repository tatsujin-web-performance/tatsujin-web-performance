FROM golang:1.26-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server .

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /server /app/server
CMD ["/app/server"]
