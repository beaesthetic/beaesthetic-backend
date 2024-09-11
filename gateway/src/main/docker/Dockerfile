FROM ghcr.io/graalvm/native-image:ol8-java17-22 AS builder

RUN microdnf install -y findutils

COPY ./gateway .

RUN chmod +x ./gradlew
RUN ./gradlew :nativeCompile -Pnative

FROM ubuntu:lunar-20230128

EXPOSE 8080

COPY --from=builder /gateway/build/native/nativeCompile/beaesthetic-gateway .

ENTRYPOINT ["./beaesthetic-gateway"]