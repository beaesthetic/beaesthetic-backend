package it.beaesthetic.fidelity.rest

import it.beaesthetic.fidelity.application.FidelityCardService
import it.beaesthetic.fidelity.domain.*
import it.beaesthetic.fidelity.generated.api.FidelityCardsApi
import it.beaesthetic.fidelity.generated.api.model.CreateFidelityCardRequestDto
import it.beaesthetic.fidelity.generated.api.model.FidelityCardResponseDto
import it.beaesthetic.fidelity.generated.api.model.PurchaseNofityRequestDto
import it.beaesthetic.fidelity.generated.api.model.SupportedVoucherTreatmentDto
import it.beaesthetic.fidelity.rest.ResourceMapper.toResource
import jakarta.ws.rs.core.Response
import org.jboss.resteasy.reactive.server.ServerExceptionMapper
import java.time.Instant
import java.util.*


class FidelityCardResource(
    private val fidelityCardService: FidelityCardService,
    private val fidelityCardRepository: FidelityCardRepository
) : FidelityCardsApi {

    override suspend fun createFidelityCard(
        createFidelityCardRequestDto: CreateFidelityCardRequestDto
    ): FidelityCardResponseDto {
        return fidelityCardService.createFidelityCard(
            CustomerId(createFidelityCardRequestDto.customerId.toString())
        ).map { it.toResource() }.getOrThrow()
    }

    override suspend fun getFidelityCardById(cardId: UUID): FidelityCardResponseDto {
        return fidelityCardRepository.findById(cardId.toString())?.toResource()
            ?: throw IllegalArgumentException("Fidelity card not found")
    }

    override suspend fun getFidelityCards(): List<FidelityCardResponseDto> {
        return fidelityCardRepository.findAll().map { it.toResource() }
    }

    override suspend fun getFidelityCardsByCustomerId(customerId: UUID): List<FidelityCardResponseDto> {
        return fidelityCardRepository.findByCustomerId(CustomerId(customerId.toString()))?.toResource()
            ?.let { listOf(it) } ?: emptyList()
    }

    override suspend fun notifyPurchase(
        cardId: UUID,
        purchaseNofityRequestDto: PurchaseNofityRequestDto
    ) {
        val purchase = TreatmentPurchase(
            at = Instant.now(),
            amount = purchaseNofityRequestDto.amount.toDouble(),
            treatment = when (purchaseNofityRequestDto.treatment) {
                SupportedVoucherTreatmentDto.SOLARIUM -> FidelityTreatment.SOLARIUM
                else -> FidelityTreatment.SOLARIUM
            }
        )
        return fidelityCardService.registerPurchase(cardId.toString(), purchase)
            .map {  }
            .getOrThrow()
    }

    override suspend fun useVoucher(voucherId: UUID): FidelityCardResponseDto {
        return fidelityCardService.useVoucher(VoucherId(voucherId.toString())).map {
            it.toResource()
        }.getOrThrow()
    }

    @ServerExceptionMapper(IllegalArgumentException::class)
    fun handleIllegal(error: IllegalArgumentException): Response {
        return when {
            error.message?.contains("already exists") ?: false -> Response.status(Response.Status.CONFLICT)
                .entity(error.message).build()

            error.message?.contains("not found") ?: false -> Response.status(Response.Status.NOT_FOUND)
                .entity(error.message).build()

            else -> Response.status(Response.Status.INTERNAL_SERVER_ERROR).entity(error.message).build()
        }
    }
}