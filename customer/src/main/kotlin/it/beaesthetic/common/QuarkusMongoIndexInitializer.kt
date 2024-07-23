package it.beaesthetic.common

import io.quarkus.runtime.Startup
import jakarta.enterprise.context.ApplicationScoped
import jakarta.enterprise.inject.Instance
import kotlinx.coroutines.runBlocking
import org.jboss.logging.Logger

@ApplicationScoped
class QuarkusMongoIndexInitializer(private val mongoInitializers: Instance<MongoInitializer>) {

    companion object {
        private val log: Logger = Logger.getLogger(QuarkusMongoIndexInitializer::class.java)
    }

    @Startup
    fun initializeIndexes() = runBlocking {
        log.info("Ensuring mongodb indexes")
        mongoInitializers.toList().forEach { it.initialize() }
    }
}
