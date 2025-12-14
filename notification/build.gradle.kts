import com.diffplug.spotless.LineEnding
import org.jetbrains.kotlin.gradle.tasks.KotlinCompile
import org.openapitools.generator.gradle.plugin.tasks.GenerateTask

plugins {
  kotlin("jvm") version "1.9.25"
  kotlin("plugin.allopen") version "1.9.25"
  id("io.quarkus")
  id("org.openapi.generator") version "7.17.0"
  id("com.diffplug.spotless") version "6.25.0"
  kotlin("kapt") version "1.9.25"
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
  implementation("io.arrow-kt:arrow-core:1.2.4")

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
  implementation("io.smallrye.reactive:mutiny-kotlin:2.9.5")
  implementation("com.fasterxml.jackson.module:jackson-module-kotlin")

  // vertx-lang-kotlin-coroutines
  implementation("io.quarkus:quarkus-vertx")
  implementation("io.vertx:vertx-lang-kotlin-coroutines:4.5.22")

  // rabbitmq
  implementation("io.quarkus:quarkus-messaging-rabbitmq")

  // validation
  implementation("io.quarkus:quarkus-hibernate-validator")

  // health check
  implementation("io.quarkus:quarkus-smallrye-health")

  // observability
  implementation("io.quarkus:quarkus-opentelemetry")
  implementation("io.quarkus:quarkus-micrometer")

  implementation(kotlin("reflect"))

  // rest client generator
  implementation("io.quarkiverse.openapi.generator:quarkus-openapi-generator:2.13.0-lts")
  implementation("io.quarkus:quarkus-rest-client-reactive-jackson")

  testImplementation("io.quarkus:quarkus-junit5")
}

group = "it.beaesthetic.customer"

version = "${properties["version"]}"

java {
  sourceCompatibility = JavaVersion.VERSION_17
  targetCompatibility = JavaVersion.VERSION_17
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

sourceSets {
  main {
    java {
      srcDirs("$buildDir/generated/src/main/java")
      srcDirs("$buildDir/classes/java/quarkus-generated-sources/open-api-yaml")
    }
  }
}

tasks.withType<KotlinCompile> {
  dependsOn("notification-api", "sms-gateway-webhook-api")
  kotlinOptions { jvmTarget = "17" }
}

configure<com.diffplug.gradle.spotless.SpotlessExtension> {
  lineEndings = LineEnding.UNIX
  kotlin {
    toggleOffOn()
    targetExclude("build/**/*")
    ktfmt().kotlinlangStyle()
    trimTrailingWhitespace()
    endWithNewline()
  }
  kotlinGradle {
    toggleOffOn()
    targetExclude("build/**/*.kts")
    ktfmt().googleStyle()
  }
}

tasks.register<GenerateTask>("notification-api") {
  description = "Generate REST API interface for customer"
  group = "openapi-generation"
  generatorName.set("kotlin-server")
  inputSpec.set("$rootDir/api-spec/notification-api.yaml")
  outputDir.set("$buildDir/generated")
  apiPackage.set("it.beaesthetic.notification.generated.api")
  modelPackage.set("it.beaesthetic.notification.generated.api.model")
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
      "NotificationChannelDto" to "it.beaesthetic.notification.driver.rest.dto.ChannelMixin",
    )
  )
  typeMappings.putAll(
    mapOf(
      "NotificationChannelDto" to "ChannelMixin",
    )
  )
}

tasks.register<GenerateTask>("sms-gateway-webhook-api") {
  description = "Generate REST API interface for customer"
  group = "openapi-generation"
  generatorName.set("kotlin-server")
  inputSpec.set("$rootDir/api-spec/sms-gateway-webhook.yaml")
  outputDir.set("$buildDir/generated")
  apiPackage.set("it.beaesthetic.notification.sms.generated.api")
  modelPackage.set("it.beaesthetic.notification.sms.generated.api.model")
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
}
