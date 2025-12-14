package it.beaesthetic.wallet.application

import it.beaesthetic.wallet.domain.*
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.assertThrows
import org.mockito.kotlin.*

class WalletServiceTest {

    private val walletRepository: WalletRepository = mock()
    private val service = WalletService(walletRepository)

    private val customerId = "customer-123"

    @Test
    fun `should create new wallet with gift card when wallet does not exist`(): Unit = runBlocking {
        // Given
        val amount = Money(100.0)

        whenever(walletRepository.findByCustomer(customerId)).thenReturn(null)
        doAnswer { invocation -> invocation.getArgument<Wallet>(0) }
            .whenever(walletRepository)
            .save(any())

        // When
        val result = service.createWallet(customerId, amount)

        // Then
        assertTrue(result.isSuccess)
        val wallet = result.getOrThrow()
        assertEquals(customerId, wallet.owner)
        assertEquals(amount, wallet.availableAmount)
        assertEquals(1, wallet.giftCards.size)
        assertEquals(amount, wallet.giftCards.first().availableAmount)

        verify(walletRepository).findByCustomer(customerId)
        verify(walletRepository).save(any())
    }

    @Test
    fun `should add gift card to existing wallet`(): Unit = runBlocking {
        // Given
        val existingWallet = Wallet.ofCustomer(customerId)
        val amount = Money(50.0)

        whenever(walletRepository.findByCustomer(customerId)).thenReturn(existingWallet)
        doAnswer { invocation -> invocation.getArgument<Wallet>(0) }
            .whenever(walletRepository)
            .save(any())

        // When
        val result = service.createWallet(customerId, amount)

        // Then
        assertTrue(result.isSuccess)
        val wallet = result.getOrThrow()
        assertEquals(amount, wallet.availableAmount)
        assertEquals(1, wallet.giftCards.size)

        verify(walletRepository).findByCustomer(customerId)
        verify(walletRepository).save(any())
    }

    @Test
    fun `should create gift card with one year expiration`(): Unit = runBlocking {
        // Given
        val amount = Money(100.0)

        whenever(walletRepository.findByCustomer(customerId)).thenReturn(null)
        doAnswer { invocation -> invocation.getArgument<Wallet>(0) }
            .whenever(walletRepository)
            .save(any())

        // When
        val result = service.createWallet(customerId, amount)

        // Then
        assertTrue(result.isSuccess)
        val wallet = result.getOrThrow()
        val giftCard = wallet.giftCards.first()

        assertNotNull(giftCard.expiresAt)
        assertEquals(customerId, giftCard.owner)

        // Gift card should expire approximately 365 days from creation
        val daysUntilExpiry =
            java.time.Duration.between(giftCard.createdAt, giftCard.expiresAt).toDays()
        assertEquals(365, daysUntilExpiry)
    }

    @Test
    fun `should record GiftCardMoneyCredited operation when creating wallet`(): Unit = runBlocking {
        // Given
        val amount = Money(100.0)

        whenever(walletRepository.findByCustomer(customerId)).thenReturn(null)
        doAnswer { invocation -> invocation.getArgument<Wallet>(0) }
            .whenever(walletRepository)
            .save(any())

        // When
        val result = service.createWallet(customerId, amount)

        // Then
        assertTrue(result.isSuccess)
        val wallet = result.getOrThrow()
        assertEquals(1, wallet.operations.size)

        val operation = wallet.operations.first()
        assertTrue(operation is GiftCardMoneyCredited)
        val creditOperation = operation as GiftCardMoneyCredited
        assertEquals(amount, creditOperation.amount)
        assertEquals(wallet.giftCards.first().id, creditOperation.giftCardId)
    }

    @Test
    fun `should charge money from wallet successfully`(): Unit = runBlocking {
        // Given
        val walletId = WalletId("wallet-123")
        val existingWallet = Wallet.ofCustomer(customerId).let { it.creditMoney(Money(100.0)) }
        val chargeAmount = Money(50.0)

        whenever(walletRepository.findById(walletId)).thenReturn(existingWallet)
        doAnswer { invocation -> invocation.getArgument<Wallet>(0) }
            .whenever(walletRepository)
            .save(any())

        // When
        val result = service.charge(walletId, chargeAmount)

        // Then
        assertTrue(result.isSuccess)
        val wallet = result.getOrThrow()
        assertEquals(Money(50.0), wallet.availableAmount)
        assertEquals(Money(50.0), wallet.spentAmount)

        verify(walletRepository).findById(walletId)
        verify(walletRepository).save(any())
    }

    @Test
    fun `should fail to charge when wallet does not exist`(): Unit = runBlocking {
        // Given
        val walletId = WalletId("non-existent")
        val chargeAmount = Money(50.0)

        whenever(walletRepository.findById(walletId)).thenReturn(null)

        // When
        val result = service.charge(walletId, chargeAmount)

        // Then
        assertTrue(result.isFailure)
        assertEquals("Wallet $walletId not found", result.exceptionOrNull()?.message)

        verify(walletRepository).findById(walletId)
        verify(walletRepository, never()).save(any())
    }

    @Test
    fun `should fail to charge when wallet has insufficient balance`(): Unit = runBlocking {
        // Given
        val walletId = WalletId("wallet-123")
        val existingWallet = Wallet.ofCustomer(customerId).let { it.creditMoney(Money(30.0)) }
        val chargeAmount = Money(50.0)

        whenever(walletRepository.findById(walletId)).thenReturn(existingWallet)

        // When
        assertThrows<IllegalArgumentException> { service.charge(walletId, chargeAmount) }
    }

    @Test
    fun `should record MoneyCharge operation when charging wallet`(): Unit = runBlocking {
        // Given
        val walletId = WalletId("wallet-123")
        val existingWallet = Wallet.ofCustomer(customerId).let { it.creditMoney(Money(100.0)) }
        val chargeAmount = Money(30.0)

        whenever(walletRepository.findById(walletId)).thenReturn(existingWallet)
        doAnswer { invocation -> invocation.getArgument<Wallet>(0) }
            .whenever(walletRepository)
            .save(any())

        // When
        val result = service.charge(walletId, chargeAmount)

        // Then
        assertTrue(result.isSuccess)
        val wallet = result.getOrThrow()

        val chargeOperation = wallet.operations.first()
        assertTrue(chargeOperation is MoneyCharge)
        assertEquals(chargeAmount, (chargeOperation as MoneyCharge).amount)
    }

    @Test
    fun `should propagate repository save failure when creating wallet`(): Unit = runBlocking {
        // Given
        val amount = Money(100.0)
        val error = RuntimeException("Database error")

        whenever(walletRepository.findByCustomer(customerId)).thenReturn(null)
        doReturn(Result.failure<Wallet>(error)).whenever(walletRepository).save(any())

        // When
        val result = service.createWallet(customerId, amount)

        // Then
        assertTrue(result.isFailure)
        assertEquals(error, result.exceptionOrNull())
    }

    @Test
    fun `should propagate repository save failure when charging wallet`(): Unit = runBlocking {
        // Given
        val walletId = WalletId("wallet-123")
        val existingWallet = Wallet.ofCustomer(customerId).let { it.creditMoney(Money(100.0)) }
        val chargeAmount = Money(50.0)
        val error = RuntimeException("Database error")

        whenever(walletRepository.findById(walletId)).thenReturn(existingWallet)
        doReturn(Result.failure<Wallet>(error)).whenever(walletRepository).save(any())

        // When
        val result = service.charge(walletId, chargeAmount)

        // Then
        assertTrue(result.isFailure)
        assertEquals(error, result.exceptionOrNull())
    }

    @Test
    fun `should add multiple gift cards to same wallet`(): Unit = runBlocking {
        // Given
        var wallet = Wallet.ofCustomer(customerId)

        whenever(walletRepository.findByCustomer(customerId)).thenAnswer { wallet }
        doAnswer { invocation ->
                wallet = invocation.getArgument<Wallet>(0)
                wallet
            }
            .whenever(walletRepository)
            .save(any())

        // When - First gift card
        service.createWallet(customerId, Money(50.0))

        // When - Second gift card
        val result = service.createWallet(customerId, Money(75.0))

        // Then
        assertTrue(result.isSuccess)
        val finalWallet = result.getOrThrow()
        assertEquals(Money(125.0), finalWallet.availableAmount)
        assertEquals(2, finalWallet.giftCards.size)
    }
}
