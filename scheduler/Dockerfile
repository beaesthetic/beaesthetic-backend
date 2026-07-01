FROM golang:1.23-alpine AS build
WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum* ./
RUN go mod download
RUN go install github.com/magefile/mage@v1.15.0

COPY . .
RUN mage build

FROM alpine:3.20
WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=build /app/scheduler /app/scheduler
COPY --from=build /app/migrations /app/migrations

EXPOSE 8080

ENTRYPOINT ["/app/scheduler"]
CMD ["app"]