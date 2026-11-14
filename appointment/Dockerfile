# syntax=docker/dockerfile:1

FROM golang:1.27.4-bookworm AS build

WORKDIR /src/appointment
COPY core /src/core
COPY core-contracts /src/core-contracts
COPY appointment/go.mod appointment/go.sum ./
RUN go mod download

COPY appointment .
RUN go run github.com/magefile/mage@v1.15.0 -v build

FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /app
COPY --from=build /src/appointment/build/appointment /app/appointment
COPY --from=build /src/appointment/migrations /app/migrations

EXPOSE 8080
ENTRYPOINT ["/app/appointment"]
CMD ["app"]
