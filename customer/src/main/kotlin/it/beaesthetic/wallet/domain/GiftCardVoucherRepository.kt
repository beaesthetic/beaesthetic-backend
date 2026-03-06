package it.beaesthetic.wallet.domain

interface GiftCardVoucherRepository {
    suspend fun save(card: GiftCardVoucher): Result<GiftCardVoucher>
    suspend fun findById(id: String): GiftCardVoucher?
    suspend fun findByCode(code: String): GiftCardVoucher?
    suspend fun findAll(): List<GiftCardVoucher>
    suspend fun findPage(limit: Int, offset: Int): List<GiftCardVoucher>
    suspend fun delete(id: String)
}
