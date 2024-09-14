FROM ghcr.io/graalvm/native-image-community:17.0.8 AS builder
WORKDIR /code

RUN microdnf install -y findutils

COPY ./scheduler .

RUN chmod +x ./gradlew
RUN ./gradlew :nativeCompile -DskipTests

FROM debian:stable-20240701-slim
WORKDIR /app/

EXPOSE 8080

COPY --from=builder /code/build/native/nativeCompile/code ./scheduler

ENTRYPOINT ["./scheduler"]