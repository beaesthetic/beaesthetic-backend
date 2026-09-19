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
COPY --from=build /src/build/identity /app/identity
COPY --from=build /src/migrations /app/migrations
COPY --from=build /src/seeds /app/seeds

EXPOSE 8080 9090
ENTRYPOINT ["/app/identity"]
CMD ["app"]
