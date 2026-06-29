FROM golang:1.23-alpine AS builder
WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/scheduler ./cmd/scheduler

FROM alpine:3.20
WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/scheduler /app/scheduler
COPY migrations /app/migrations

EXPOSE 8080

ENTRYPOINT ["/app/scheduler"]
CMD ["app"]
