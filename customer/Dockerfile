FROM golang:1.25-bookworm AS build

WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN go run github.com/magefile/mage@v1.15.0 -v build

FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /app
COPY --from=build /src/build/customer /app/customer
COPY --from=build /src/migrations /app/migrations

EXPOSE 8080
ENTRYPOINT ["/app/customer"]
CMD ["app"]