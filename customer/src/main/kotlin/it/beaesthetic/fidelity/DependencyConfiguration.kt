package it.beaesthetic.fidelity

import com.mongodb.client.model.IndexModel
import com.mongodb.client.model.Indexes
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.common.MongoInitializer
import it.beaesthetic.fidelity.application.FidelityCardService
import it.beaesthetic.fidelity.domain.FidelityCardRepository
import it.beaesthetic.fidelity.infra.PanacheFidelityCardRepository
import jakarta.enterprise.context.Dependent
import jakarta.enterprise.inject.Produces

@Dependent
class DependencyConfiguration {

    @Produces
    fun fidelityCardService(fidelityCardRepository: FidelityCardRepository): FidelityCardService {
        return FidelityCardService(fidelityCardRepository)
    }

    @Produces
    fun mongoInitializer(
        panacheFidelityCardRepository: PanacheFidelityCardRepository,
    ): MongoInitializer =
        object : MongoInitializer {
            override suspend fun initialize() {
                panacheFidelityCardRepository
                    .mongoCollection()
                    .createIndexes(listOf(IndexModel(Indexes.descending("vouchers.id"))))
                    .awaitSuspending()
            }
        }
}
