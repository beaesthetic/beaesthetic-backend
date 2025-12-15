import com.diffplug.spotless.LineEnding
import org.jetbrains.kotlin.gradle.tasks.KotlinCompile
import org.openapitools.generator.gradle.plugin.tasks.GenerateTask

plugins {
  kotlin("jvm") version "2.2.21"
  kotlin("plugin.spring") version "2.2.21"
  kotlin("plugin.allopen") version "2.2.21"
  id("org.springframework.boot") version "4.0.0"
  id("io.spring.dependency-management") version "1.1.7"
  id("org.graalvm.buildtools.native") version "0.11.3"
  id("org.openapi.generator") version "7.17.0"
  id("com.diffplug.spotless") version "8.1.0"
  kotlin("kapt") version "2.2.21"
}

repositories {
  mavenCentral()
  mavenLocal()
}

configurations {
  implementation.configure {
    exclude(module = "spring-boot-starter-web")
    exclude("org.apache.tomcat")
    exclude(group = "org.slf4j", module = "slf4j-simple")
  }
}

dependencies {
  // functional - fp
  implementation("io.arrow-kt:arrow-core:2.2.0")

  // kotlin
  implementation(kotlin("stdlib-jdk8"))
  implementation("com.fasterxml.jackson.module:jackson-module-kotlin")

  // spring boot
  implementation("org.springframework.boot:spring-boot-starter-webflux")
  implementation("io.projectreactor.kotlin:reactor-kotlin-extensions")
  implementation("org.jetbrains.kotlinx:kotlinx-coroutines-reactor")
  implementation("org.springframework.boot:spring-boot-starter-validation")
  implementation("org.springframework.boot:spring-boot-starter-amqp") // rabbit
  implementation("org.springframework.boot:spring-boot-starter-actuator") // actuator

  // redis
  implementation("org.springframework.boot:spring-boot-starter-data-redis-reactive")

  implementation(kotlin("reflect"))

  // jakarta
  implementation("jakarta.xml.bind:jakarta.xml.bind-api")

  // openapi
  implementation("org.openapitools:openapi-generator-gradle-plugin:7.14.0")
  implementation("org.openapitools:jackson-databind-nullable:0.2.8")
  implementation("io.swagger.core.v3:swagger-annotations:2.2.41")

  testImplementation("org.springframework.boot:spring-boot-starter-actuator-test")
  testImplementation("org.springframework.boot:spring-boot-starter-amqp-test")
  testImplementation("org.springframework.boot:spring-boot-starter-data-redis-reactive-test")
  testImplementation("org.springframework.boot:spring-boot-starter-webflux-test")
  testImplementation("org.jetbrains.kotlin:kotlin-test-junit5")
  testImplementation("io.projectreactor:reactor-test")
  testImplementation("org.jetbrains.kotlinx:kotlinx-coroutines-test")
  testRuntimeOnly("org.junit.platform:junit-platform-launcher")
}

group = "io.github.petretiandrea.scheduler"

version = "${properties["version"]}"

java {
  sourceCompatibility = JavaVersion.VERSION_21
  targetCompatibility = JavaVersion.VERSION_21
  toolchain { languageVersion.set(JavaLanguageVersion.of(21)) }
}

allOpen {}

sourceSets {
  main {
    kotlin {
      srcDirs("${layout.buildDirectory.get()}/generated/src/main/kotlin")
      // srcDirs("${layout.buildDirectory.get()}/classes/java/quarkus-generated-sources/open-api")
    }
    java { srcDirs("${layout.buildDirectory.get()}/generated/src/main/java") }
  }
}

kotlin {
  compilerOptions {
    jvmTarget = org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_21
    javaParameters = true
    freeCompilerArgs.addAll("-Xjsr305=strict")
  }
  jvmToolchain { languageVersion.set(JavaLanguageVersion.of(21)) }
}

tasks.withType<KotlinCompile> { dependsOn("scheduler-api") }

spotless {
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

tasks.register<GenerateTask>("scheduler-api") {
  description = "Generate Scheduler API interface"
  group = "openapi-generation"

  generatorName.set("kotlin-spring")
  inputSpec.set("$rootDir/api-spec/openapi.yaml")
  outputDir.set("${layout.buildDirectory.get()}/generated")
  apiPackage.set("io.github.petretiandrea.scheduler.api")
  modelPackage.set("io.github.petretiandrea.scheduler.model")
  generateApiTests.set(false)
  generateApiDocumentation.set(false)
  generateApiTests.set(false)
  generateModelTests.set(false)
  library.set("spring-boot")
  modelNameSuffix.set("Dto")
  configOptions.set(
    mapOf(
      "documentationProvider" to "none",
      "interfaceOnly" to "true",
      "skipDefaultInterface" to "true",
      "useSwaggerUI" to "false",
      "useSpringBoot3" to "true",
      "useBeanValidation" to "true",
      "enumPropertyNaming" to "UPPERCASE",
      "useJakartaEe" to "false",
      "reactive" to "true",
      "exceptionHandler" to "false",
    )
  )
}
