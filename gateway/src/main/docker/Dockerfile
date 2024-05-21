FROM ghcr.io/graalvm/native-image:ol8-java17-22 AS builder

RUN microdnf install -y findutils

COPY . .

RUN chmod +x ./gradlew
RUN ./gradlew :gateway:nativeCompile -Pnative

FROM ubuntu:lunar-20230128

EXPOSE 8080

COPY --from=builder /app/gateway/build/native/nativeCompile/gateway .

ENTRYPOINT ["./gateway"]