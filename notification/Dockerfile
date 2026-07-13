# syntax=docker/dockerfile:1

FROM golang:1.25-bookworm AS build

WORKDIR /src
COPY go.mod go.sum* ./
COPY --from=core-contracts / /core-contracts
RUN go mod download

COPY . .
RUN go run github.com/magefile/mage@v1.15.0 -v build

FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /app
COPY --from=build /src/build/notification /app/notification
COPY --from=build /src/migrations /app/migrations

EXPOSE 8080
ENTRYPOINT ["/app/notification"]
CMD ["app"]
