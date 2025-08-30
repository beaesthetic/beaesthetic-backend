package it.beaesthetic.fidelity.infra

import com.mongodb.client.model.Aggregates.*
import com.mongodb.client.model.Filters.eq
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.fidelity.application.FidelityCardReadService
import it.beaesthetic.fidelity.application.data.FidelityCardReadModel
import it.beaesthetic.fidelity.domain.CustomerId
import jakarta.enterprise.context.ApplicationScoped

@ApplicationScoped
class FidelityCardReadServiceImpl(
    private val panacheFidelityCardRepository: PanacheFidelityCardRepository
) : FidelityCardReadService {

    private val mongoCollection = panacheFidelityCardRepository.mongoCollection()

    private val joinCustomerPipeline =
        listOf(
            lookup("customers", "customerId", "_id", "customer"),
            unwind("\$customer"),
            unset("customerId"),
        )

    override suspend fun findAll(): List<FidelityCardReadModel> {
        return mongoCollection
            .aggregate(joinCustomerPipeline, FidelityCardReadModel::class.java)
            .collect()
            .asList()
            .awaitSuspending()
    }

    override suspend fun findById(id: String): FidelityCardReadModel? {
        val pipeline = listOf(match(eq("_id", id))) + joinCustomerPipeline

        return mongoCollection
            .aggregate(pipeline, FidelityCardReadModel::class.java)
            .collect()
            .asList()
            .awaitSuspending()
            .firstOrNull()
    }

    override suspend fun findByCustomerId(customerId: CustomerId): FidelityCardReadModel? {
        val pipeline = listOf(match(eq("customerId", customerId.id))) + joinCustomerPipeline

        return mongoCollection
            .aggregate(pipeline, FidelityCardReadModel::class.java)
            .collect()
            .asList()
            .awaitSuspending()
            .firstOrNull()
    }
}
