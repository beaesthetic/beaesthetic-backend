FROM ghcr.io/graalvm/native-image:ol8-java17-22 AS builder

RUN microdnf install -y findutils

COPY ./gateway .

RUN chmod +x ./gradlew
RUN ./gradlew :nativeCompile -Pnative

FROM ubuntu:24.04

EXPOSE 8080

COPY --from=builder /app/build/native/nativeCompile/beaesthetic-gateway .

ENTRYPOINT ["./beaesthetic-gateway"]