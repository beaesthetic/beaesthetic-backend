package it.beaesthetic.customer

import com.fasterxml.jackson.databind.ObjectMapper
import com.mongodb.client.model.IndexModel
import com.mongodb.client.model.Indexes
import io.quarkus.mongodb.reactive.ReactiveMongoClient
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.common.MongoInitializer
import it.beaesthetic.customer.application.CustomerService
import it.beaesthetic.customer.domain.CustomerEvent
import it.beaesthetic.customer.domain.CustomerRepository
import it.beaesthetic.customer.infra.OutboxRepository
import it.beaesthetic.customer.infra.PanacheCustomerRepository
import jakarta.enterprise.context.Dependent
import jakarta.enterprise.inject.Produces
import org.eclipse.microprofile.config.inject.ConfigProperty

@Dependent
class DependencyConfiguration {

    @Produces
    fun customerService(customerRepository: CustomerRepository): CustomerService {
        return CustomerService(customerRepository)
    }

    @ConfigProperty(name = "quarkus.mongodb.database") lateinit var databaseName: String

    @Produces
    fun outboxRepository(
        mongoClient: ReactiveMongoClient,
        objectMapper: ObjectMapper,
    ): OutboxRepository<CustomerEvent> =
        OutboxRepository(mongoClient.getDatabase(databaseName), objectMapper, "outboxitems")

    @Produces
    fun mongoInitializer(panacheCustomerRepository: PanacheCustomerRepository): MongoInitializer =
        object : MongoInitializer {
            override suspend fun initialize() {
                panacheCustomerRepository
                    .mongoCollection()
                    .createIndexes(
                        listOf(
                            IndexModel(Indexes.ascending("name")),
                            IndexModel(Indexes.ascending("surname")),
                            IndexModel(Indexes.text("searchGrams")),
                        )
                    )
                    .awaitSuspending()
            }
        }
}
