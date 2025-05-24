package it.beaesthetic.wallet.rest

import it.beaesthetic.wallet.application.WalletService
import it.beaesthetic.wallet.domain.*
import it.beaesthetic.wallet.generated.api.WalletsApi
import it.beaesthetic.wallet.generated.api.model.*
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
        return walletRepository.findById(WalletId(walletId.toString()))?.toResource()
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

    private fun Wallet.toResource() =
        WalletDto(
            id = UUID.fromString(id),
            customerId = UUID.fromString(owner),
            availableAmount = availableAmount.amount.toBigDecimal(),
            spent = spentAmount.amount.toBigDecimal(),
            history = operations.map { op -> op.toResource() },
            createdAt = createdAt.atOffset(ZoneOffset.UTC),
            updatedAt = updatedAt.atOffset(ZoneOffset.UTC),
        )

    private fun WalletEvent.toResource() =
        when (this) {
            is GiftCardMoneyCredited ->
                GiftCardMoneyCreditedEventDto(
                    giftCardId = UUID.fromString(giftCardId),
                    at = at.atOffset(ZoneOffset.UTC),
                    amount = amount.amount.toBigDecimal(),
                    expireAt = expiresAt.atOffset(ZoneOffset.UTC),
                )
            is GiftCardMoneyExpired ->
                GiftCardMoneyExpiredEventDto(
                    giftCardId = UUID.fromString(giftCardId),
                    at = at.atOffset(ZoneOffset.UTC),
                    amount = 0.toBigDecimal(),
                )
            is MoneyCharge ->
                MoneyChargedEventDto(
                    amount = amount.amount.toBigDecimal(),
                    at = at.atOffset(ZoneOffset.UTC),
                )
            is MoneyCredited ->
                MoneyCreditedEventDto(
                    amount = amount.amount.toBigDecimal(),
                    at = at.atOffset(ZoneOffset.UTC),
                )
        }
}
