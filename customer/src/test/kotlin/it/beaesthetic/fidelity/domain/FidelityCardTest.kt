package it.beaesthetic.fidelity.domain

import java.time.Instant
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertNotNull
import kotlin.test.assertNull
import kotlin.test.assertTrue

class FidelityCardTest {

    private val customerId = CustomerId("customer-123")
    private val now = Instant.parse("2024-01-15T10:00:00Z")

    @Test
    fun `should create FidelityCard for customer with empty vouchers`() {
        val card = FidelityCard.ofCustomer(customerId)

        assertNotNull(card.id)
        assertEquals(customerId, card.customerId)
        assertTrue(card.availableVouchers.isEmpty())
        assertEquals(0, card.purchasesOf(FidelityTreatment.SOLARIUM))
    }

    @Test
    fun `should increment purchase counter when adding purchase`() {
        val card = FidelityCard.ofCustomer(customerId)
        val purchase = TreatmentPurchase(now, 50.0, FidelityTreatment.SOLARIUM)

        val voucher = card.addPurchase(purchase)

        assertNull(voucher)
        assertEquals(1, card.purchasesOf(FidelityTreatment.SOLARIUM))
        assertTrue(card.availableVouchers.isEmpty())
    }

    @Test
    fun `should generate voucher after 10 purchases of same treatment`() {
        val card = FidelityCard.ofCustomer(customerId)
        var voucher: Voucher? = null

        // Add 9 purchases - no voucher
        for (i in 1..9) {
            val purchase = TreatmentPurchase(now, 50.0, FidelityTreatment.SOLARIUM)
            voucher = card.addPurchase(purchase)
            assertNull(voucher)
        }

        // 10th purchase - voucher generated
        val tenthPurchase = TreatmentPurchase(now, 50.0, FidelityTreatment.SOLARIUM)
        voucher = card.addPurchase(tenthPurchase)

        assertNotNull(voucher)
        assertFalse(voucher.isUsed)
        assertEquals(FidelityTreatment.SOLARIUM, voucher.treatment)
        assertEquals(1, card.availableVouchers.size)
        assertEquals(0, card.purchasesOf(FidelityTreatment.SOLARIUM)) // Counter resets
    }

    @Test
    fun `should reset purchase counter to zero after generating voucher`() {
        val card = FidelityCard.ofCustomer(customerId)

        // Add 10 purchases to generate voucher
        for (i in 1..10) {
            val purchase = TreatmentPurchase(now, 50.0, FidelityTreatment.SOLARIUM)
            card.addPurchase(purchase)
        }

        assertEquals(0, card.purchasesOf(FidelityTreatment.SOLARIUM))

        // Add one more purchase - counter should be 1
        val nextPurchase = TreatmentPurchase(now, 50.0, FidelityTreatment.SOLARIUM)
        card.addPurchase(nextPurchase)

        assertEquals(1, card.purchasesOf(FidelityTreatment.SOLARIUM))
    }

    @Test
    fun `should generate multiple vouchers with multiple rounds of 10 purchases`() {
        val card = FidelityCard.ofCustomer(customerId)

        // First 10 purchases
        for (i in 1..10) {
            val purchase = TreatmentPurchase(now, 50.0, FidelityTreatment.SOLARIUM)
            card.addPurchase(purchase)
        }
        assertEquals(1, card.availableVouchers.size)

        // Second 10 purchases
        for (i in 1..10) {
            val purchase = TreatmentPurchase(now, 50.0, FidelityTreatment.SOLARIUM)
            card.addPurchase(purchase)
        }
        assertEquals(2, card.availableVouchers.size)
    }

    @Test
    fun `should use voucher successfully when voucher exists and is not used`() {
        val card = FidelityCard.ofCustomer(customerId)

        // Generate a voucher
        for (i in 1..10) {
            val purchase = TreatmentPurchase(now, 50.0, FidelityTreatment.SOLARIUM)
            card.addPurchase(purchase)
        }

        val voucher = card.availableVouchers.first()
        val result = card.useVoucher(voucher.id)

        assertTrue(result.isSuccess)
        val updatedCard = result.getOrThrow()
        val usedVoucher = updatedCard.availableVouchers.first { it.id == voucher.id }
        assertTrue(usedVoucher.isUsed)
    }

    @Test
    fun `should fail to use voucher when voucher does not exist`() {
        val card = FidelityCard.ofCustomer(customerId)
        val nonExistentVoucherId = VoucherId("non-existent")

        val result = card.useVoucher(nonExistentVoucherId)

        assertTrue(result.isFailure)
        assertEquals("No voucher found", result.exceptionOrNull()?.message)
    }

    @Test
    fun `should fail to use voucher when voucher is already used`() {
        val card = FidelityCard.ofCustomer(customerId)

        // Generate and use a voucher
        for (i in 1..10) {
            val purchase = TreatmentPurchase(now, 50.0, FidelityTreatment.SOLARIUM)
            card.addPurchase(purchase)
        }

        val voucher = card.availableVouchers.first()
        val usedCard = card.useVoucher(voucher.id).getOrThrow()

        // Try to use again
        val usedVoucher = usedCard.availableVouchers.first { it.id == voucher.id }
        val result = usedCard.useVoucher(usedVoucher.id)

        assertTrue(result.isFailure)
        assertTrue(result.exceptionOrNull()?.message!!.contains("already used"))
    }

    @Test
    fun `should fail to use voucher object when already marked as used`() {
        val card = FidelityCard.ofCustomer(customerId)

        // Generate a voucher
        for (i in 1..10) {
            val purchase = TreatmentPurchase(now, 50.0, FidelityTreatment.SOLARIUM)
            card.addPurchase(purchase)
        }

        val voucher = card.availableVouchers.first()
        val usedCard = card.useVoucher(voucher.id).getOrThrow()
        val usedVoucher = usedCard.availableVouchers.first { it.id == voucher.id }

        val result = usedCard.useVoucher(usedVoucher)

        assertTrue(result.isFailure)
        assertEquals("Voucher already used", result.exceptionOrNull()?.message)
    }

    @Test
    fun `should return all vouchers in availableVouchers including used ones`() {
        val card = FidelityCard.ofCustomer(customerId)

        // Generate two vouchers
        for (i in 1..20) {
            val purchase = TreatmentPurchase(now, 50.0, FidelityTreatment.SOLARIUM)
            card.addPurchase(purchase)
        }

        assertEquals(2, card.availableVouchers.size)

        // Use first voucher
        val firstVoucher = card.availableVouchers.first()
        val updatedCard = card.useVoucher(firstVoucher.id).getOrThrow()

        // Both vouchers should still be in the list
        assertEquals(2, updatedCard.availableVouchers.size)
        assertEquals(1, updatedCard.availableVouchers.count { it.isUsed })
        assertEquals(1, updatedCard.availableVouchers.count { !it.isUsed })
    }

    @Test
    fun `should return zero for purchases of treatment with no purchases`() {
        val card = FidelityCard.ofCustomer(customerId)

        assertEquals(0, card.purchasesOf(FidelityTreatment.SOLARIUM))
    }

    @Test
    fun `should create voucher with correct treatment type`() {
        val card = FidelityCard.ofCustomer(customerId)

        for (i in 1..10) {
            val purchase = TreatmentPurchase(now, 50.0, FidelityTreatment.SOLARIUM)
            card.addPurchase(purchase)
        }

        val voucher = card.availableVouchers.first()
        assertEquals(FidelityTreatment.SOLARIUM, voucher.treatment)
    }

    @Test
    fun `should maintain purchase counter independently after voucher generation`() {
        val card = FidelityCard.ofCustomer(customerId)

        // Add 15 purchases total
        for (i in 1..15) {
            val purchase = TreatmentPurchase(now, 50.0, FidelityTreatment.SOLARIUM)
            card.addPurchase(purchase)
        }

        // Should have 1 voucher and counter at 5
        assertEquals(1, card.availableVouchers.size)
        assertEquals(5, card.purchasesOf(FidelityTreatment.SOLARIUM))
    }
}
