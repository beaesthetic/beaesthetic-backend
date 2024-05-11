package it.beaesthetic.insights

import io.quarkus.mongodb.MongoClientName
import io.quarkus.mongodb.reactive.ReactiveMongoClient
import it.beaesthetic.insights.repository.TreatmentCustomerCountRepository
import it.beaesthetic.insights.repository.TreatmentUsageCountRepository
import it.beaesthetic.insights.service.InsightQueryService
import it.beaesthetic.insights.service.InsightService
import jakarta.enterprise.context.Dependent
import jakarta.enterprise.inject.Produces

@Dependent
class AppConfiguration {

    @Produces
    fun insightService(
        @MongoClientName("appointment") appointment: ReactiveMongoClient,
        treatmentUsageCountRepository: TreatmentUsageCountRepository,
        treatmentCustomerCountRepository: TreatmentCustomerCountRepository
    ): InsightService {
        return InsightService(appointment, treatmentUsageCountRepository, treatmentCustomerCountRepository)
    }

    @Produces
    fun queryInsightsService(
        treatmentUsageCountRepository: TreatmentUsageCountRepository,
        treatmentCustomerCountRepository: TreatmentCustomerCountRepository
    ): InsightQueryService = InsightQueryService(
        treatmentUsageCountRepository,
        treatmentCustomerCountRepository
    )

}