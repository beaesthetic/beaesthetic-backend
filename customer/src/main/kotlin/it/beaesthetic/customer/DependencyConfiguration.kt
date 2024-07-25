package it.beaesthetic.customer

import com.mongodb.client.model.IndexModel
import com.mongodb.client.model.IndexOptions
import com.mongodb.client.model.Indexes
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.common.MongoInitializer
import it.beaesthetic.customer.application.CustomerService
import it.beaesthetic.customer.domain.CustomerRepository
import it.beaesthetic.customer.infra.PanacheCustomerRepository
import jakarta.enterprise.context.Dependent
import jakarta.enterprise.inject.Produces

@Dependent
class DependencyConfiguration {

    @Produces
    fun customerService(customerRepository: CustomerRepository): CustomerService {
        return CustomerService(customerRepository)
    }

    @Produces
    fun mongoInitializer(
        panacheCustomerRepository: PanacheCustomerRepository,
    ): MongoInitializer =
        object : MongoInitializer {
            override suspend fun initialize() {
                panacheCustomerRepository
                    .mongoCollection()
                    .createIndexes(
                        listOf(
                            IndexModel(Indexes.text("searchGrams"))
                        )
                    )
                    .awaitSuspending()
            }
        }
}
