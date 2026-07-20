import org.graalvm.buildtools.gradle.tasks.BuildNativeImageTask
import org.jetbrains.kotlin.gradle.tasks.KotlinCompile
import com.github.jengelman.gradle.plugins.shadow.tasks.ShadowJar

plugins {
    id("org.springframework.boot") version "3.2.5"
    id("io.spring.dependency-management") version "1.1.4"
    kotlin("jvm") version "1.9.23"
    kotlin("plugin.spring") version "1.9.23"

    id("org.graalvm.buildtools.native") version "0.9.28"
    id("com.github.johnrengelman.shadow") version "7.1.2"
}

group = "it.beaesthetic.gateway"
version = "${project.version}"

java {
    sourceCompatibility = JavaVersion.VERSION_17
}

repositories {
    mavenCentral()
    maven { url = uri("https://repo.spring.io/milestone") }
    maven { url = uri("https://repo.spring.io/snapshot") }
}

springBoot {
    mainClass.set("it.beaesthetic.gateway.GatewayApplicationKt")
}

extra["springCloudVersion"] = "2023.0.6"

dependencies {
    // functional - fp
    implementation("io.arrow-kt:arrow-core:1.1.2")

    // spring boot
    implementation("org.springframework.cloud:spring-cloud-starter-gateway")
    implementation("org.springframework.boot:spring-boot-starter-actuator")
    implementation("org.springframework.boot:spring-boot-starter-security")
    implementation("org.springframework.boot:spring-boot-starter-webflux")

    // jackson kotlin
    implementation("com.fasterxml.jackson.module:jackson-module-kotlin")
    implementation("io.projectreactor.kotlin:reactor-kotlin-extensions")
    implementation("org.jetbrains.kotlin:kotlin-reflect")
    implementation("org.jetbrains.kotlin:kotlin-stdlib-jdk8")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.6.4")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-guava:1.6.4")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-reactor")

    // gson
    implementation("com.google.code.gson:gson:2.10.1")

    // firebase
    implementation("com.google.firebase:firebase-admin:9.2.0")

    // https://mvnrepository.com/artifact/io.grpc/grpc-netty
    implementation("io.grpc:grpc-netty")

    // jwt
    implementation("io.jsonwebtoken:jjwt-api:0.11.5")
    runtimeOnly("io.jsonwebtoken:jjwt-impl:0.11.5")
    runtimeOnly("io.jsonwebtoken:jjwt-jackson:0.11.5")

    // otel
    implementation("io.opentelemetry.instrumentation:opentelemetry-spring-boot-starter")
    implementation("io.opentelemetry:opentelemetry-exporter-otlp")

    testImplementation("org.springframework.boot:spring-boot-starter-test")
    testImplementation("io.projectreactor:reactor-test")
    testImplementation("org.springframework.security:spring-security-test")
    testRuntimeOnly("org.junit.platform:junit-platform-launcher")
}

dependencyManagement {
    imports {
        mavenBom("org.springframework.cloud:spring-cloud-dependencies:${property("springCloudVersion")}")
        mavenBom("io.grpc:grpc-bom:1.59.0")
        mavenBom("io.opentelemetry:opentelemetry-bom:1.38.0")
        mavenBom("io.opentelemetry.instrumentation:opentelemetry-instrumentation-bom-alpha:2.4.0-alpha")
    }
}

graalvmNative {
    metadataRepository {
        enabled.set(true)
        version.set("0.2.6")
    }

    binaries {
        useArgFile.set(false)
        named("main") {
            useFatJar.set(true)
            javaLauncher.set(javaToolchains.launcherFor {
                languageVersion.set(JavaLanguageVersion.of(17))
                vendor.set(JvmVendorSpec.GRAAL_VM)
            })
            mainClass.set("it.beaesthetic.gateway.GatewayApplicationKt")
            buildArgs.add("--add-opens=java.base/java.nio=ALL-UNNAMED")
            buildArgs.add("--add-opens=java.base/jdk.internal.misc=ALL-UNNAMED")
            buildArgs.add("--add-opens=java.base/jdk.internal.ref=ALL-UNNAMED")
            buildArgs.add("--initialize-at-run-time=io.netty")
        }
    }
}

tasks.withType<KotlinCompile> {
    kotlinOptions {
        freeCompilerArgs += "-Xjsr305=strict"
        jvmTarget = "17"
    }
}

tasks.withType<Test> {
    useJUnitPlatform()
}

tasks.named<ShadowJar>("shadowJar") {
    isZip64 = true
    archiveBaseName.set("shadow")
    duplicatesStrategy = DuplicatesStrategy.WARN
    mergeServiceFiles()
    manifest {
        attributes(mapOf("Main-Class" to "it.beaesthetic.gateway.GatewayApplicationKt"))
    }

    from(graalvmNative.binaries.named("main").get().classpath
        .elements.map { elem ->
            elem.stream().map { if (it.asFile.path.endsWith(".jar")) zipTree(it) else it }.toList()
        })
}

tasks.named<BuildNativeImageTask>("nativeCompile") {
    val shadowJarTask = tasks.getByName<ShadowJar>("shadowJar")
    dependsOn(shadowJarTask)
    classpathJar.set(shadowJarTask.archiveFile.get())
}