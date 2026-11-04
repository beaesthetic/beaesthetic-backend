FROM golang:1.27.4-bookworm AS build

WORKDIR /src/customer
COPY core /src/core
COPY customer/go.mod customer/go.sum ./
RUN go mod download

COPY customer .
RUN go run github.com/magefile/mage@v1.15.0 -v build

FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /app
COPY --from=build /src/customer/build/customer /app/customer
COPY --from=build /src/customer/migrations /app/migrations

EXPOSE 8080
ENTRYPOINT ["/app/customer"]
CMD ["app"]
