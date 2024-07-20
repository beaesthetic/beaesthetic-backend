package it.beaesthetic.common

import com.mongodb.client.model.IndexModel
import com.mongodb.client.model.IndexOptions
import com.mongodb.client.model.Indexes
import io.quarkus.runtime.Startup
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.customer.infra.PanacheCustomerRepository
import jakarta.enterprise.context.ApplicationScoped
import kotlinx.coroutines.runBlocking
import org.jboss.logging.Logger

@ApplicationScoped
class QuarkusMongoIndexInitializer(
    private val panacheCustomerRepository: PanacheCustomerRepository
) {

    companion object {
        private val log: Logger = Logger.getLogger(QuarkusMongoIndexInitializer::class.java)
    }

    @Startup
    fun initializeIndexes() = runBlocking {
        log.info("Ensuring mongodb indexes")
        panacheCustomerRepository.mongoCollection()
            .createIndexes(listOf(
                IndexModel(Indexes.descending("id"), IndexOptions().unique(true)),
                IndexModel(Indexes.text("searchGrams"))
            )).awaitSuspending()
    }

}