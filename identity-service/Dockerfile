FROM golang:1.25-bookworm AS build

WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /identity ./cmd/identity

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /identity /identity
EXPOSE 8080
ENTRYPOINT ["/identity"]
