# see https://quarkus.io/guides/building-native-image#multistage-docker
FROM quay.io/quarkus/ubi-quarkus-mandrel-builder-image@sha256:eaeefa7e92980ca6ff13572808153ce5b55d3efe08f40ee15bdf5109a68ac2b8 AS builder
USER root
RUN microdnf install findutils
#COPY --chown=quarkus:quarkus --chmod=0755 gradle.lockfile /code/
COPY --chown=quarkus:quarkus --chmod=0755 ./ /code/
#COPY --chown=quarkus:quarkus --chmod=0755 gradle.properties /code/
USER quarkus
WORKDIR /code
RUN ./gradlew customer:build -Dquarkus.package.type=native -DskipTests

# also a distroless is available https://quarkus.io/guides/building-native-image#using-a-distroless-base-image
FROM quay.io/quarkus/quarkus-micro-image:2.0 AS runtime
WORKDIR /work/
RUN chown 1001 /work \
    && chmod "g+rwX" /work \
    && chown 1001:root /work
COPY --from=builder --chown=1001:root --chmod=0755 /code/customer/build/*-runner /work/application

EXPOSE 8080
USER 1001

ENTRYPOINT ["./application", "-Dquarkus.http.host=0.0.0.0"]