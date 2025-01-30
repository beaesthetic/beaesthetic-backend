import com.diffplug.spotless.LineEnding
import org.jetbrains.kotlin.gradle.tasks.KotlinCompile
import org.openapitools.generator.gradle.plugin.tasks.GenerateTask

plugins {
  kotlin("jvm") version "2.0.21"
  kotlin("plugin.allopen") version "2.0.21"
  id("io.quarkus")
  id("org.openapi.generator") version "7.5.0"
  id("com.diffplug.spotless") version "6.18.0"
  kotlin("kapt") version "2.0.21"
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
  implementation("io.smallrye.reactive:mutiny-kotlin:2.8.0")
  implementation("com.fasterxml.jackson.module:jackson-module-kotlin")

  // vertx-lang-kotlin-coroutines
  implementation("io.quarkus:quarkus-vertx")
  implementation("io.vertx:vertx-lang-kotlin-coroutines:4.5.7")

  // redis
  implementation("io.quarkus:quarkus-redis-client")
  implementation("io.quarkus:quarkus-redis-cache")

  // rabbitmq
  implementation("io.quarkus:quarkus-messaging-rabbitmq")

  // validation
  implementation("io.quarkus:quarkus-hibernate-validator")

  // health check
  implementation("io.quarkus:quarkus-smallrye-health")

  // observability
  implementation("io.quarkus:quarkus-opentelemetry")
  implementation("io.quarkus:quarkus-micrometer")
  implementation("io.opentelemetry.instrumentation:opentelemetry-micrometer-1.5")

  // rest client generator
  implementation("io.quarkiverse.openapi.generator:quarkus-openapi-generator:2.8.0")
  implementation("io.quarkus:quarkus-rest-client-reactive-jackson")

  implementation(kotlin("reflect"))

  testImplementation("io.quarkus:quarkus-junit5")
  testImplementation(kotlin("test"))
  testImplementation("org.mockito.kotlin:mockito-kotlin:5.4.0")
}

group = "it.beaesthetic.appointment"

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
  dependsOn("appointment-api")
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

tasks.register<GenerateTask>("appointment-api") {
  description = "Generate REST API interface for customer"
  group = "openapi-generation"
  generatorName.set("kotlin-server")
  inputSpec.set("$rootDir/api-spec/openapi.yaml")
  outputDir.set("$buildDir/generated")
  apiPackage.set("it.beaesthetic.appointment.agenda.generated.api")
  modelPackage.set("it.beaesthetic.appointment.agenda.generated.api.model")
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
      "useMutiny" to "true",
      "useCoroutines" to "false",
      "useBeanValidation" to "true",
      "dateLibrary" to "java8",
      "useJakartaEe" to "true",
      "useTags" to "true",
      "returnResponse" to "false"
    )
  )
  typeMappings.putAll(
    mapOf(
      "CreateAgendaActivityRequestDto" to "CreateAgendaActivityMixin",
    )
  )
  schemaMappings.putAll(mapOf("CreateAgendaActivityRequest" to "CreateAgendaActivityMixin"))

  importMappings.putAll(
    mapOf(
      "CreateAgendaActivityRequestDto" to
        "it.beaesthetic.appointment.agenda.port.rest.CreateAgendaActivityMixin",
      "CreateAgendaActivityMixin" to
        "it.beaesthetic.appointment.agenda.port.rest.CreateAgendaActivityMixin"
    )
  )
}
