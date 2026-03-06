package it.beaesthetic.wallet.http

import it.beaesthetic.giftcard.generated.api.GiftCardsAdminV2Api
import it.beaesthetic.giftcard.generated.api.model.CreateGiftCardRequestDto
import it.beaesthetic.giftcard.generated.api.model.GiftCardDto
import it.beaesthetic.giftcard.generated.api.model.GiftCardPageDto
import it.beaesthetic.giftcard.generated.api.model.GiftCardStatusDto
import it.beaesthetic.giftcard.generated.api.model.RedeemGiftCardRequestDto
import it.beaesthetic.giftcard.generated.api.model.RedemptionResultDto
import it.beaesthetic.wallet.application.WalletService
import it.beaesthetic.wallet.domain.Money
import it.beaesthetic.wallet.domain.GiftCardVoucher
import jakarta.ws.rs.core.Response
import jakarta.ws.rs.core.UriInfo
import java.math.BigDecimal
import java.net.URI
import java.time.Instant
import java.time.ZoneOffset
import java.util.UUID
import com.fasterxml.jackson.annotation.JsonProperty
import jakarta.inject.Inject
import jakarta.ws.rs.core.Context
import org.jboss.resteasy.reactive.server.ServerExceptionMapper

class GiftCardV2Resource(
    private val walletService: WalletService,
) : GiftCardsAdminV2Api {

    @Context
    lateinit var uriInfo: UriInfo

    override suspend fun createGiftCard(createGiftCardRequestDto: CreateGiftCardRequestDto): Response {
        val amount = Money(createGiftCardRequestDto.amount.toDouble())
        return walletService.createGiftCardVoucher(amount)
            .map { card ->
                val location = uriInfo.absolutePathBuilder
                    .path(card.id)
                    .build()
                Response.created(location)
                    .entity(card.toDto())
                    .build()
            }
            .getOrThrow()
    }

    override suspend fun listGiftCards(
        status: GiftCardStatusDto?,
        code: String?,
        limit: Int?,
        cursor: String?,
    ): GiftCardPageDto {
        val pageSize = (limit ?: 20).coerceIn(1, 100)
        val offset = cursor?.let { decodeCursor(it) } ?: 0
        val now = Instant.now()

        val page = if (code != null) {
            // Lookup by code — returns at most one item
            listOfNotNull(walletService.getGiftCardVouchers().firstOrNull { it.code == code })
        } else {
            walletService.listGiftCards(limit = pageSize + 1, offset = offset)
        }

        val hasMore = code == null && page.size > pageSize
        val items = if (hasMore) page.dropLast(1) else page

        val filteredItems = when (status) {
            GiftCardStatusDto.EXPIRED -> items.filter { it.isExpiredAt(now) }
            GiftCardStatusDto.PENDING -> items.filter { !it.isExpiredAt(now) }
            null -> items
        }

        val nextCursor = if (hasMore) encodeCursor(offset + pageSize) else null

        return GiftCardPageDto(
            items = filteredItems.map { it.toDto(now) },
            nextCursor = nextCursor,
        )
    }

    override suspend fun getGiftCard(giftCardId: UUID): GiftCardDto {
        return walletService.getGiftCard(giftCardId.toString())?.toDto()
            ?: throw jakarta.ws.rs.NotFoundException("Gift card not found")
    }

    override suspend fun redeemGiftCard(
        giftCardId: UUID,
        redeemGiftCardRequestDto: RedeemGiftCardRequestDto,
    ): RedemptionResultDto {
        return walletService.redeemGiftCardById(
            id = giftCardId.toString(),
            customerId = redeemGiftCardRequestDto.customerId.toString(),
        )
            .map { wallet -> RedemptionResultDto(walletId = UUID.fromString(wallet.id)) }
            .getOrThrow()
    }

    // ---- Mapping ----

    private fun GiftCardVoucher.toDto(now: Instant = Instant.now()) =
        GiftCardDto(
            id = UUID.fromString(id),
            code = code,
            amount = amount.amount.toBigDecimal(),
            status = if (isExpiredAt(now)) GiftCardStatusDto.EXPIRED else GiftCardStatusDto.PENDING,
            createdAt = createdAt.atOffset(ZoneOffset.UTC),
            expiresAt = expiresAt.atOffset(ZoneOffset.UTC),
        )

    // ---- Exception mapping ----

    @ServerExceptionMapper(IllegalArgumentException::class)
    fun handleIllegal(error: IllegalArgumentException): Response {
        val problem = when {
            error.message?.contains("not found", ignoreCase = true) == true ->
                GiftCardProblem.notFound(error.message)
            error.message?.contains("expired", ignoreCase = true) == true ->
                GiftCardProblem.expired(error.message)
            error.message?.contains("amount", ignoreCase = true) == true ->
                GiftCardProblem.invalidAmount(error.message)
            else ->
                GiftCardProblem.badRequest(error.message)
        }
        return Response.status(problem.status)
            .type("application/problem+json")
            .entity(problem)
            .build()
    }

    // ---- Cursor helpers (opaque base-10 offset) ----


    private fun encodeCursor(offset: Int): String =
        java.util.Base64.getUrlEncoder().encodeToString(offset.toString().toByteArray())

    private fun decodeCursor(cursor: String): Int =
        runCatching {
            java.util.Base64.getUrlDecoder().decode(cursor).toString(Charsets.UTF_8).toInt()
        }.getOrDefault(0)
}

data class GiftCardProblem(
    @JsonProperty("type") val type: String,
    @JsonProperty("title") val title: String,
    @JsonProperty("status") val status: Int,
    @JsonProperty("detail") val detail: String?,
) {
    companion object {
        fun notFound(detail: String?) = GiftCardProblem(
            type = "/problems/gift-card-not-found",
            title = "Gift Card Not Found",
            status = 404,
            detail = detail,
        )

        fun expired(detail: String?) = GiftCardProblem(
            type = "/problems/gift-card-expired",
            title = "Gift Card Expired",
            status = 422,
            detail = detail,
        )

        fun invalidAmount(detail: String?) = GiftCardProblem(
            type = "/problems/invalid-amount",
            title = "Invalid Amount",
            status = 422,
            detail = detail,
        )

        fun badRequest(detail: String?) = GiftCardProblem(
            type = "/problems/bad-request",
            title = "Bad Request",
            status = 400,
            detail = detail,
        )
    }
}
