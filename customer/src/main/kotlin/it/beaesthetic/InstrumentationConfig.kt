package it.beaesthetic

import com.mongodb.MongoClientSettings
import io.opentelemetry.api.OpenTelemetry
import io.opentelemetry.instrumentation.mongo.v3_1.MongoTelemetry
import io.quarkus.mongodb.runtime.MongoClientCustomizer
import jakarta.enterprise.context.Dependent
import jakarta.enterprise.inject.Produces

@Dependent
class InstrumentationConfig {
    @Produces
    fun mongoOtelCustomization(openTelemetry: OpenTelemetry): MongoClientCustomizer =
        OtelMongoCustomizer(openTelemetry)

    inner class OtelMongoCustomizer(private val openTelemetry: OpenTelemetry) :
        MongoClientCustomizer {
        override fun customize(builder: MongoClientSettings.Builder): MongoClientSettings.Builder? {
            val mongoTelemetry = MongoTelemetry.builder(openTelemetry).build()
            return builder.addCommandListener(mongoTelemetry.newCommandListener())
        }
    }
}
