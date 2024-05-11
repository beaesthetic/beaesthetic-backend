package it.beaesthetic.insights.adapter

import io.quarkus.scheduler.Scheduled
import it.beaesthetic.insights.service.InsightService
import org.jboss.logging.Logger
import java.time.Instant

class ScheduledInsights(
    private val insightService: InsightService,
) {

    private val logger = Logger.getLogger(ScheduledInsights::class.java.name)

    @Scheduled(cron = "{insights.batch.cron}")
    suspend fun computeInsights() {
        try {
            logger.info("Start computing insights...")
            insightService.computeMostUsedTreatments(Instant.now())
            insightService.computeTreatmentsByCustomer(Instant.now())
            logger.info("Computing insights completed")
        } catch (error: Exception) {
            logger.error("Error during computing insights", error)
        }
    }

}