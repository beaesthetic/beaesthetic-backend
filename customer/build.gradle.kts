import org.openapitools.generator.gradle.plugin.tasks.GenerateTask

plugins {
  kotlin("jvm") version "1.9.22"
  kotlin("plugin.allopen") version "1.9.22"
  id("io.quarkus")
  id("org.openapi.generator") version "7.5.0"
  id("com.diffplug.spotless") version "6.18.0"
}

repositories {
  mavenCentral()
  mavenLocal()
}

val quarkusPlatformGroupId: String by project
val quarkusPlatformArtifactId: String by project
val quarkusPlatformVersion: String by project

dependencies {
  // functional - fp
  implementation("io.arrow-kt:arrow-core:1.1.2")

  // quarkus
  implementation(
    enforcedPlatform(
      "${quarkusPlatformGroupId}:${quarkusPlatformArtifactId}:${quarkusPlatformVersion}"
    )
  )
  implementation("io.quarkus:quarkus-resteasy-reactive")
  implementation("io.quarkus:quarkus-resteasy-reactive-jackson")
  implementation("io.quarkus:quarkus-smallrye-health")
  implementation("io.quarkus:quarkus-mongodb-client")
  implementation("io.quarkus:quarkus-arc")

  // quarkus & kotlin
  implementation("io.quarkus:quarkus-kotlin")
  implementation("org.jetbrains.kotlin:kotlin-stdlib-jdk8")
  implementation("io.quarkus:quarkus-mongodb-panache-kotlin")
  implementation("io.smallrye.reactive:mutiny-kotlin:2.6.0")
  implementation("com.fasterxml.jackson.module:jackson-module-kotlin")

  // vertx-lang-kotlin-coroutines
  implementation("io.vertx:vertx-lang-kotlin-coroutines:4.5.7")

  // validation
  implementation("io.quarkus:quarkus-hibernate-validator")

  // redis
  implementation("io.quarkus:quarkus-redis-client")
  implementation("io.quarkus:quarkus-redis-cache")

  // health check
  implementation("io.quarkus:quarkus-smallrye-health")

  // observability
  implementation("io.quarkus:quarkus-opentelemetry")
  implementation("io.quarkus:quarkus-micrometer")

  implementation(kotlin("reflect"))

  testImplementation("io.quarkus:quarkus-junit5")
}

group = "it.beaesthetic.customer"

version

"${properties["version"]}"

java {
  sourceCompatibility = JavaVersion.VERSION_21
  targetCompatibility = JavaVersion.VERSION_21
}

tasks.withType<Test> {
  systemProperty("java.util.logging.manager", "org.jboss.logmanager.LogManager")
}

allOpen {
  annotation("jakarta.ws.rs.Path")
  annotation("jakarta.enterprise.context.ApplicationScoped")
  annotation("jakarta.persistence.Entity")
  annotation("io.quarkus.test.junit.QuarkusTest")
}

sourceSets { main { java { srcDirs("$buildDir/generated/src/main/java") } } }

tasks.register<GenerateTask>("customer-api") {
  description = "Generate REST API interface for customer"
  group = "openapi-generation"
  generatorName.set("kotlin-server")
  inputSpec.set("$rootDir/api-spec/customer-api.yaml")
  outputDir.set("$buildDir/generated")
  apiPackage.set("it.beaesthetic.customer.generated.api")
  modelPackage.set("it.beaesthetic.customer.generated.api.model")
  generateApiTests.set(false)
  generateApiDocumentation.set(false)
  generateApiTests.set(false)
  generateModelTests.set(false)
  validateSpec.set(true)

  library.set("jaxrs-spec")
  modelNameSuffix.set("Dto")
  configOptions.set(
    mapOf(
      "sourceFolder" to "src/main/java",
      "skipDefaultInterface" to "true",
      "openApiNullable" to "true",
      "hideGenerationTimestamp" to "true",
      "oas3" to "true",
      "generateSupportingFiles" to "true",
      "enumPropertyNaming" to "UPPERCASE",
      "legacyDiscriminatorBehavior" to "true",
      "interfaceOnly" to "true",
      "useSwaggerAnnotations" to "false",
      "supportAsync" to "true",
      "useMutiny" to "true",
      "useCoroutines" to "false",
      "useBeanValidation" to "true",
      "dateLibrary" to "java8",
      "useJakartaEe" to "true",
      "useTags" to "true"
    )
  )
}

tasks.register<GenerateTask>("fidelity-card-api") {
  description = "Generate REST API interface for customer"
  group = "openapi-generation"
  generatorName.set("kotlin-server")
  inputSpec.set("$rootDir/api-spec/fidelity-card-api.yaml")
  outputDir.set("$buildDir/generated")
  apiPackage.set("it.beaesthetic.fidelity.generated.api")
  modelPackage.set("it.beaesthetic.fidelity.generated.api.model")
  generateApiTests.set(false)
  generateApiDocumentation.set(false)
  generateApiTests.set(false)
  generateModelTests.set(false)
  validateSpec.set(true)

  library.set("jaxrs-spec")
  modelNameSuffix.set("Dto")
  templateDir.set("${projectDir.path}/src/main/resources/kotlin-server")

  configOptions.set(
    mapOf(
      "sourceFolder" to "src/main/java",
      "skipDefaultInterface" to "true",
      "openApiNullable" to "true",
      "hideGenerationTimestamp" to "true",
      "oas3" to "true",
      "generateSupportingFiles" to "true",
      "enumPropertyNaming" to "UPPERCASE",
      "legacyDiscriminatorBehavior" to "true",
      "interfaceOnly" to "true",
      "useSwaggerAnnotations" to "false",
      "supportAsync" to "true",
      "useMutiny" to "false",
      "useCoroutines" to "true",
      "useBeanValidation" to "true",
      "dateLibrary" to "java8",
      "useJakartaEe" to "true",
      "useTags" to "true"
    )
  )

  importMappings.putAll(
    mapOf(
      "VoucherDto" to "it.beaesthetic.fidelity.rest.serialization.VoucherItemMixin",
    )
  )
  typeMappings.putAll(
    mapOf(
      "VoucherDto" to "VoucherItemMixin",
    )
  )
}

configure<com.diffplug.gradle.spotless.SpotlessExtension> {
  kotlin {
    toggleOffOn()
    targetExclude("build/**/*")
    ktfmt().kotlinlangStyle()
  }
  kotlinGradle {
    toggleOffOn()
    targetExclude("build/**/*.kts")
    ktfmt().googleStyle()
  }
}
