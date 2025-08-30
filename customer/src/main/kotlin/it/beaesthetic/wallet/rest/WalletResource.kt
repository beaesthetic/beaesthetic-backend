package it.beaesthetic.wallet.rest

import it.beaesthetic.wallet.application.WalletService
import it.beaesthetic.wallet.application.read.WalletReadModel
import it.beaesthetic.wallet.domain.*
import it.beaesthetic.wallet.generated.api.WalletsApi
import it.beaesthetic.wallet.generated.api.model.*
import it.beaesthetic.wallet.infra.*
import jakarta.ws.rs.NotFoundException
import jakarta.ws.rs.core.Response
import java.time.ZoneOffset
import java.util.*
import org.jboss.resteasy.reactive.server.ServerExceptionMapper

class WalletResource(
    private val walletService: WalletService,
    private val walletRepository: WalletRepository,
) : WalletsApi {

    override suspend fun addGiftCard(
        addGiftCardRequestDto: AddGiftCardRequestDto
    ): AddGiftCard200ResponseDto {
        return walletService
            .createWallet(
                addGiftCardRequestDto.customerId.toString(),
                Money(addGiftCardRequestDto.amount.toDouble()),
            )
            .map { AddGiftCard200ResponseDto(id = UUID.fromString(it.id)) }
            .getOrThrow()
    }

    override suspend fun chargeWallet(
        walletId: UUID,
        chargeWalletRequestDto: ChargeWalletRequestDto,
    ): AddGiftCard200ResponseDto {
        return walletService
            .charge(WalletId(walletId.toString()), Money(chargeWalletRequestDto.amount.toDouble()))
            .map { AddGiftCard200ResponseDto(id = UUID.fromString(it.id)) }
            .getOrThrow()
    }

    override suspend fun getWalletById(walletId: UUID): WalletDto {
        return walletRepository.findByIdReadModel(WalletId(walletId.toString()))?.toResource()
            ?: throw NotFoundException("Wallet not found")
    }

    override suspend fun getWallets(filter: String?): List<WalletDto> {
        return walletRepository.findAll().map { it.toResource() }
    }

    @ServerExceptionMapper(IllegalArgumentException::class)
    fun handleIllegal(error: IllegalArgumentException): Response {
        return when {
            error.message?.contains("already exists") ?: false ->
                Response.status(Response.Status.CONFLICT).entity(error.message).build()
            error.message?.contains("not found") ?: false ->
                Response.status(Response.Status.NOT_FOUND).entity(error.message).build()
            error.message?.contains("cause exceed maximum amount of") ?: false ->
                Response.status(422).entity(error.message).build()
            else ->
                Response.status(Response.Status.INTERNAL_SERVER_ERROR).entity(error.message).build()
        }
    }

    private fun WalletReadModel.toResource() =
        WalletDto(
            id = UUID.fromString(wallet.id),
            customer =
                CustomerDto(
                    id = customer.id,
                    name = customer.name ?: "",
                    surname = customer.surname ?: "",
                    phone = customer.phone,
                    email = customer.email,
                ),
            availableAmount = wallet.availableAmount.toBigDecimal(),
            spent = wallet.spentAmount.toBigDecimal(),
            history = wallet.operations.map { op -> op.toResource() },
            createdAt = wallet.createdAt.atOffset(ZoneOffset.UTC),
            updatedAt = wallet.updatedAt.atOffset(ZoneOffset.UTC),
        )

    private fun WalletEventEntity.toResource() =
        when (this) {
            is GiftCardMoneyCreditedEntity ->
                GiftCardMoneyCreditedEventDto(
                    giftCardId = UUID.fromString(giftCardId),
                    at = at.atOffset(ZoneOffset.UTC),
                    amount = amount.toBigDecimal(),
                    expireAt = expiresAt.atOffset(ZoneOffset.UTC),
                )
            is GiftCardMoneyExpiredEntity ->
                GiftCardMoneyExpiredEventDto(
                    giftCardId = UUID.fromString(giftCardId),
                    at = at.atOffset(ZoneOffset.UTC),
                    amount = 0.toBigDecimal(),
                )
            is MoneyChargedEntity ->
                MoneyChargedEventDto(
                    amount = amount.toBigDecimal(),
                    at = at.atOffset(ZoneOffset.UTC),
                )
            is MoneyCreditedEntity ->
                MoneyCreditedEventDto(
                    amount = amount.toBigDecimal(),
                    at = at.atOffset(ZoneOffset.UTC),
                )
        }
}
