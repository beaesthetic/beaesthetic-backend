# see https://quarkus.io/guides/building-native-image#multistage-docker
FROM quay.io/quarkus/ubi-quarkus-mandrel-builder-image:25.0-jdk-25 AS builder
USER root
RUN microdnf install findutils
#COPY --chown=quarkus:quarkus --chmod=0755 gradle.lockfile /code/
COPY --chown=quarkus:quarkus --chmod=0755 ./appointment /code/
#COPY --chown=quarkus:quarkus --chmod=0755 gradle.properties /code/
USER quarkus
WORKDIR /code
RUN ./gradlew :build -Dquarkus.package.type=native -DskipTests

# also a distroless is available https://quarkus.io/guides/building-native-image#using-a-distroless-base-image
FROM quay.io/quarkus/quarkus-micro-image:2.0 AS runtime
WORKDIR /work/
RUN chown 1001 /work \
    && chmod "g+rwX" /work \
    && chown 1001:root /work
COPY --from=builder --chown=1001:root --chmod=0755 /code/build/*-runner /work/application

EXPOSE 8080
USER 1001

ENTRYPOINT ["./application", "-Dquarkus.http.host=0.0.0.0"]