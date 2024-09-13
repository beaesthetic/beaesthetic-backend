import com.diffplug.spotless.LineEnding
import org.jetbrains.kotlin.gradle.tasks.KotlinCompile
import org.openapitools.generator.gradle.plugin.tasks.GenerateTask

plugins {
  kotlin("jvm") version "1.9.22"
  kotlin("plugin.spring") version "1.9.22"
  kotlin("plugin.allopen") version "1.9.22"
  id("org.springframework.boot") version "3.3.2"
  id("io.spring.dependency-management") version "1.1.6"
  id("org.graalvm.buildtools.native") version "0.10.2"
  id("org.openapi.generator") version "7.5.0"
  id("com.diffplug.spotless") version "6.18.0"
  kotlin("kapt") version "1.7.0"
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
  implementation("io.arrow-kt:arrow-core:1.1.2")

  // kotlin
  implementation("org.jetbrains.kotlin:kotlin-stdlib-jdk8")
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
  implementation("org.openapitools:openapi-generator-gradle-plugin:6.5.0")
  implementation("org.openapitools:jackson-databind-nullable:0.2.6")
  implementation("io.swagger.core.v3:swagger-annotations:2.2.22")

  testImplementation("org.springframework.boot:spring-boot-starter-test")
  testImplementation("org.springframework.amqp:spring-rabbit-test")
  testImplementation("org.springframework.boot:spring-boot-starter-test")
  testImplementation("io.projectreactor:reactor-test")
  testImplementation("org.jetbrains.kotlin:kotlin-test-junit5")
  testRuntimeOnly("org.junit.platform:junit-platform-launcher")
}

group = "io.github.petretiandrea.scheduler"

version = "${properties["version"]}"

java { toolchain { languageVersion = JavaLanguageVersion.of(17) } }

allOpen {}

sourceSets {
  main {
    java {
      kotlin { srcDirs("src/main/kotlin", "$buildDir/generated/src/main/kotlin") }
      java { srcDirs("$buildDir/generated/src/main/java") }
    }
  }
}

kotlin { compilerOptions { freeCompilerArgs.addAll("-Xjsr305=strict") } }

tasks.withType<KotlinCompile> {
  dependsOn("scheduler-api")
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

tasks.register<GenerateTask>("scheduler-api") {
  description = "Generate Scheduler API interface"
  group = "openapi-generation"

  generatorName.set("kotlin-spring")
  inputSpec.set("$rootDir/api-spec/openapi.yaml")
  outputDir.set("$buildDir/generated")
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
      "exceptionHandler" to "false"
    )
  )
}
