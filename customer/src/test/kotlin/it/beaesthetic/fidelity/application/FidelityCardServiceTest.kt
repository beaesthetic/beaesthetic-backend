package it.beaesthetic.fidelity.application

import it.beaesthetic.fidelity.domain.*
import java.time.Instant
import kotlin.test.*
import kotlinx.coroutines.runBlocking
import org.mockito.kotlin.*

class FidelityCardServiceTest {

    private val fidelityCardRepository: FidelityCardRepository = mock()
    private val service = FidelityCardService(fidelityCardRepository)

    private val customerId = CustomerId("customer-123")
    private val now = Instant.parse("2024-01-15T10:00:00Z")

    @Test
    fun `should create fidelity card for customer`(): Unit = runBlocking {
        // Given
        whenever(fidelityCardRepository.findByCustomerId(customerId)).thenReturn(null)
        doAnswer { invocation -> invocation.getArgument<FidelityCard>(0) }
            .whenever(fidelityCardRepository)
            .save(any())

        // When
        val result = service.createFidelityCard(customerId)

        // Then
        assertTrue(result.isSuccess)
        val card = result.getOrThrow()
        assertEquals(customerId, card.customerId)
        assertTrue(card.availableVouchers.isEmpty())
        assertEquals(0, card.purchasesOf(FidelityTreatment.SOLARIUM))

        verify(fidelityCardRepository).findByCustomerId(customerId)
        verify(fidelityCardRepository).save(any())
    }

    @Test
    fun `should fail to create fidelity card when customer already has one`(): Unit = runBlocking {
        // Given
        val existingCard = FidelityCard.ofCustomer(customerId)

        whenever(fidelityCardRepository.findByCustomerId(customerId)).thenReturn(existingCard)

        // When
        val result = service.createFidelityCard(customerId)

        // Then
        assertTrue(result.isFailure)
        assertEquals("Fidelity card already exists", result.exceptionOrNull()?.message)

        verify(fidelityCardRepository).findByCustomerId(customerId)
        verify(fidelityCardRepository, never()).save(any())
    }

    @Test
    fun `should register purchase on fidelity card`(): Unit = runBlocking {
        // Given
        val cardId = "card-123"
        val card = FidelityCard.ofCustomer(customerId)
        val purchase = TreatmentPurchase(now, 50.0, FidelityTreatment.SOLARIUM)

        whenever(fidelityCardRepository.findById(cardId)).thenReturn(card)
        doAnswer { invocation -> invocation.getArgument<FidelityCard>(0) }
            .whenever(fidelityCardRepository)
            .save(any())

        // When
        val result = service.registerPurchase(cardId, purchase)

        // Then
        assertTrue(result.isSuccess)
        val voucher = result.getOrThrow()
        assertNull(voucher) // No voucher on first purchase

        verify(fidelityCardRepository).findById(cardId)
        verify(fidelityCardRepository).save(any())
    }

    @Test
    fun `should return voucher when registering 10th purchase`(): Unit = runBlocking {
        // Given
        val cardId = "card-123"
        val card = FidelityCard.ofCustomer(customerId)

        // Add 9 purchases first
        for (i in 1..9) {
            val purchase = TreatmentPurchase(now, 50.0, FidelityTreatment.SOLARIUM)
            card.addPurchase(purchase)
        }

        val tenthPurchase = TreatmentPurchase(now, 50.0, FidelityTreatment.SOLARIUM)

        whenever(fidelityCardRepository.findById(cardId)).thenReturn(card)
        doAnswer { invocation -> invocation.getArgument<FidelityCard>(0) }
            .whenever(fidelityCardRepository)
            .save(any())

        // When
        val result = service.registerPurchase(cardId, tenthPurchase)

        // Then
        assertTrue(result.isSuccess)
        val voucher = result.getOrThrow()
        assertNotNull(voucher)
        assertEquals(FidelityTreatment.SOLARIUM, voucher.treatment)
        assertFalse(voucher.isUsed)

        verify(fidelityCardRepository).findById(cardId)
        verify(fidelityCardRepository).save(any())
    }

    @Test
    fun `should fail to register purchase when fidelity card not found`(): Unit = runBlocking {
        // Given
        val cardId = "non-existent"
        val purchase = TreatmentPurchase(now, 50.0, FidelityTreatment.SOLARIUM)

        whenever(fidelityCardRepository.findById(cardId)).thenReturn(null)

        // When
        val result = service.registerPurchase(cardId, purchase)

        // Then
        assertTrue(result.isFailure)
        assertEquals("Fidelity card not found", result.exceptionOrNull()?.message)

        verify(fidelityCardRepository).findById(cardId)
        verify(fidelityCardRepository, never()).save(any())
    }

    @Test
    fun `should use voucher successfully`(): Unit = runBlocking {
        // Given
        val card = FidelityCard.ofCustomer(customerId)

        // Generate a voucher
        for (i in 1..10) {
            val purchase = TreatmentPurchase(now, 50.0, FidelityTreatment.SOLARIUM)
            card.addPurchase(purchase)
        }

        val voucher = card.availableVouchers.first()
        val voucherId = voucher.id

        whenever(fidelityCardRepository.findByVoucherId(voucherId.value)).thenReturn(card)
        doAnswer { invocation -> invocation.getArgument<FidelityCard>(0) }
            .whenever(fidelityCardRepository)
            .save(any())

        // When
        val result = service.useVoucher(voucherId)

        // Then
        assertTrue(result.isSuccess)
        val updatedCard = result.getOrThrow()
        val usedVoucher = updatedCard.availableVouchers.first { it.id == voucherId }
        assertTrue(usedVoucher.isUsed)

        verify(fidelityCardRepository).findByVoucherId(voucherId.value)
        verify(fidelityCardRepository).save(any())
    }

    @Test
    fun `should fail to use voucher when fidelity card not found`(): Unit = runBlocking {
        // Given
        val voucherId = VoucherId("non-existent")

        whenever(fidelityCardRepository.findByVoucherId(voucherId.value)).thenReturn(null)

        // When
        val result = service.useVoucher(voucherId)

        // Then
        assertTrue(result.isFailure)
        assertEquals("Fidelity card not found", result.exceptionOrNull()?.message)

        verify(fidelityCardRepository).findByVoucherId(voucherId.value)
        verify(fidelityCardRepository, never()).save(any())
    }

    @Test
    fun `should fail to use voucher when voucher does not exist on card`(): Unit = runBlocking {
        // Given
        val card = FidelityCard.ofCustomer(customerId)
        val nonExistentVoucherId = VoucherId("non-existent")

        whenever(fidelityCardRepository.findByVoucherId(nonExistentVoucherId.value))
            .thenReturn(card)

        // When
        val result = service.useVoucher(nonExistentVoucherId)

        // Then
        assertTrue(result.isFailure)

        verify(fidelityCardRepository).findByVoucherId(nonExistentVoucherId.value)
        verify(fidelityCardRepository, never()).save(any())
    }

    @Test
    fun `should fail to use voucher when voucher is already used`(): Unit = runBlocking {
        // Given
        val card = FidelityCard.ofCustomer(customerId)

        // Generate and use a voucher
        for (i in 1..10) {
            val purchase = TreatmentPurchase(now, 50.0, FidelityTreatment.SOLARIUM)
            card.addPurchase(purchase)
        }

        val voucher = card.availableVouchers.first()
        val usedCard = card.useVoucher(voucher.id).getOrThrow()
        val usedVoucher = usedCard.availableVouchers.first { it.id == voucher.id }

        whenever(fidelityCardRepository.findByVoucherId(usedVoucher.id.value)).thenReturn(usedCard)

        // When - Try to use already used voucher
        val result = service.useVoucher(usedVoucher.id)

        // Then
        assertTrue(result.isFailure)

        verify(fidelityCardRepository).findByVoucherId(usedVoucher.id.value)
        verify(fidelityCardRepository, never()).save(any())
    }

    @Test
    fun `should propagate repository save failure when creating card`(): Unit = runBlocking {
        // Given
        val error = RuntimeException("Database error")

        whenever(fidelityCardRepository.findByCustomerId(customerId)).thenReturn(null)
        doReturn(Result.failure<FidelityCard>(error)).whenever(fidelityCardRepository).save(any())

        // When
        val result = service.createFidelityCard(customerId)

        // Then
        assertTrue(result.isFailure)
        assertEquals(error, result.exceptionOrNull())
    }

    @Test
    fun `should propagate repository save failure when registering purchase`(): Unit = runBlocking {
        // Given
        val cardId = "card-123"
        val card = FidelityCard.ofCustomer(customerId)
        val purchase = TreatmentPurchase(now, 50.0, FidelityTreatment.SOLARIUM)
        val error = RuntimeException("Database error")

        whenever(fidelityCardRepository.findById(cardId)).thenReturn(card)
        doReturn(Result.failure<FidelityCard>(error)).whenever(fidelityCardRepository).save(any())

        // When
        val result = service.registerPurchase(cardId, purchase)

        // Then
        assertTrue(result.isFailure)
        assertEquals(error, result.exceptionOrNull())
    }

    @Test
    fun `should propagate repository save failure when using voucher`(): Unit = runBlocking {
        // Given
        val card = FidelityCard.ofCustomer(customerId)

        // Generate a voucher
        for (i in 1..10) {
            val purchase = TreatmentPurchase(now, 50.0, FidelityTreatment.SOLARIUM)
            card.addPurchase(purchase)
        }

        val voucher = card.availableVouchers.first()
        val error = RuntimeException("Database error")

        whenever(fidelityCardRepository.findByVoucherId(voucher.id.value)).thenReturn(card)
        doReturn(Result.failure<FidelityCard>(error)).whenever(fidelityCardRepository).save(any())

        // When
        val result = service.useVoucher(voucher.id)

        // Then
        assertTrue(result.isFailure)
        assertEquals(error, result.exceptionOrNull())
    }
}
